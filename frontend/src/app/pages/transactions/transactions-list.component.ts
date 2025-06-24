import { Component, Inject, OnInit } from '@angular/core';
import { OverlayModule } from 'primeng/overlay';
import { FormsModule } from '@angular/forms';
import { ToastModule } from 'primeng/toast';
import { TableLazyLoadEvent, TableModule } from 'primeng/table';
import { InputIcon } from 'primeng/inputicon';
import { IconField } from 'primeng/iconfield';
import { CommonModule, DatePipe } from '@angular/common';
import { Button } from 'primeng/button';
import { MultiSelectModule } from 'primeng/multiselect';
import { SelectModule } from 'primeng/select';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { FilterMetadata, MessageService } from 'primeng/api';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { ListTransactionsRequestSchema, TransactionsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { Transaction, TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { TimestampHelper } from '../../helpers/timestamp.helper';
import { AccountsService, ListAccountsResponse_AccountItem } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { AccountTypeEnum, EnumService } from '../../services/enum.service';
import { Account } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { ErrorHelper } from '../../helpers/error.helper';
import { create } from '@bufbuild/protobuf';
import { SelectedDateService } from '../../core/services/selected-date.service';
import { TimestampSchema } from '@bufbuild/protobuf/wkt';

@Component({
    selector: 'app-transaction-list',
    templateUrl: 'transactions-list.component.html',
    imports: [OverlayModule, FormsModule, ToastModule, TableModule, InputIcon, IconField, DatePipe, Button, MultiSelectModule, SelectModule, CommonModule, RouterLink]
})
export class TransactionsListComponent implements OnInit {
    private transactionsService;
    public loading = false;
    public transactions: Transaction[] = [];
    public accountsService;
    public transactionTypes: AccountTypeEnum[] = EnumService.getBaseTransactionTypes();
    public transactionTypesMap: { [id: string]: AccountTypeEnum } = {};

    public filters: { [s: string]: FilterMetadata } = {};
    public accountMap: { [id: number]: Account } = {};
    public accounts: Account[] = [];

    private currentAccountId: number | undefined;
    private ignoreDateFilter: boolean = false;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        public router: Router,
        private activeRoute: ActivatedRoute,
        private selectedDateService: SelectedDateService
    ) {
        this.transactionsService = createClient(TransactionsService, this.transport);
        this.accountsService = createClient(AccountsService, this.transport);
        this.currentAccountId = parseInt(this.activeRoute.snapshot.params['accountId']) ?? undefined;
    }

    async ngOnInit() {
        for (let type of this.transactionTypes) {
            this.transactionTypesMap[type.value] = type;
        }

        try {
            let resp = await this.accountsService.listAccounts({});
            for (let account of resp.accounts) {
                this.accountMap[account.account!.id] = account.account!;
                this.accounts.push(account.account!);
            }
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        } finally {
            this.loading = false;
        }
    }

    getAccountName(accountId: number | undefined): string {
        if (!accountId) {
            return '';
        }

        let account = this.accountMap[accountId];
        if (!account) {
            return '';
        }

        return account.name || '';
    }

    async fetchTransactions(event: TableLazyLoadEvent) {
        console.log(event);

        let req = create(ListTransactionsRequestSchema, {
            limit: event.rows ?? 50,
            skip: event.first ?? 0,
            fromDate: create(TimestampSchema, {
                seconds: BigInt(Math.floor(this.selectedDateService.getFromDate().getTime() / 1000)),
                nanos: (this.selectedDateService.getFromDate().getMilliseconds() % 1000) * 1_000_000
            }),
            toDate: create(TimestampSchema, {
                seconds: BigInt(Math.floor(this.selectedDateService.getToDate().getTime() / 1000)),
                nanos: (this.selectedDateService.getToDate().getMilliseconds() % 1000) * 1_000_000
            })
        });

        if (this.currentAccountId) {
            req.anyAccountIds = req.anyAccountIds ?? [];

            req.anyAccountIds.push(this.currentAccountId);
        }

        if (event.filters) {
            let sourceAccountIds = event.filters['sourceAccountIds'] as FilterMetadata;

            if (sourceAccountIds && sourceAccountIds.value && Array.isArray(sourceAccountIds.value)) {
                req.sourceAccountIds = sourceAccountIds.value.map((id) => parseInt(id as string));
            }

            let destinationAccountIds = event.filters['destinationAccountIds'] as FilterMetadata;
            if (destinationAccountIds && destinationAccountIds.value && Array.isArray(destinationAccountIds.value)) {
                req.destinationAccountIds = destinationAccountIds.value.map((id) => parseInt(id as string));
            }

            let title = event.filters['title'] as FilterMetadata;
            if (title && title.value && typeof title.value === 'string' && title.value.trim() !== '') {
                req.textQuery = title.value.trim();
            }
        }

        switch (event.sortField) {
            case 'transactionItem.transactionDate.nanos':
                console.log('Sorting by transaction date');
                // todo
                break;
            default:
                console.log('Unknown sort field:', event.sortField);
        }

        let resp = await this.transactionsService.listTransactions(req);

        this.transactions = resp.transactions;
    }

    getTransactionType(transaction: Transaction): string {
        if (!transaction.type) {
            return 'Unknown';
        }

        let type = this.transactionTypesMap[transaction.type];
        if (!type) {
            return 'Unknown';
        }

        return type.name;
    }

    getTransactionTypeColor(transaction: Transaction): string[] {
        return ['text-wrap', 'break-all', this.getAmountColor(transaction)];
    }
    getAmountColor(transaction: Transaction): string {
        switch (transaction.type) {
            case TransactionType.WITHDRAWAL:
                return 'text-red-500';
            case TransactionType.DEPOSIT:
                return 'text-green-500';
            case TransactionType.TRANSFER_BETWEEN_ACCOUNTS:
                return 'text-blue-500';
            default:
                return 'text-gray-500';
        }
    }

    getAccountUrl(accountId: number | undefined): string {
        if (!accountId) {
            return '';
        }

        return this.router.createUrlTree(['/', 'accounts', accountId.toString()]).toString();
    }

    getAccountColorClass(accID: number | undefined): string[] {
        let result = ['text-wrap', 'break-all'];

        if (this.currentAccountId && this.currentAccountId === accID) {
            result.push('text-purple-500');
        }

        return result;
    }

    abs(val: string | undefined): number | undefined {
        if (val === undefined) {
            return undefined;
        }

        return Math.abs(parseFloat(val));
    }

    formatAmounts(transaction: Transaction): string {
        let val = '';

        let hasSource = false;

        if (transaction.sourceAmount) {
            val += `${transaction.sourceAmount} ${transaction.sourceCurrency ?? ''}`;
            hasSource = true;
        }

        if (this.abs(transaction.sourceAmount) == this.abs(transaction.destinationAmount) && transaction.sourceCurrency == transaction.destinationCurrency) {
            return val;
        }

        if (transaction.destinationAmount) {
            if (hasSource) {
                val += ' (';
            }
            val += `${transaction.destinationAmount} ${transaction.destinationCurrency ?? ''}`;

            if (hasSource) {
                val += ')';
            }
        }

        return val.trim();
    }

    protected readonly TimestampHelper = TimestampHelper;
}
