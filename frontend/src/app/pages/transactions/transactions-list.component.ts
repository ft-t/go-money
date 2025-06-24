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
import { ActivatedRoute, Router } from '@angular/router';
import {
    ListTransactionsRequestSchema,
    TransactionsService
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { Transaction, TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { TimestampHelper } from '../../helpers/timestamp.helper';
import { AccountsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { AccountTypeEnum } from '../../services/enum.service';
import { Account } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { ErrorHelper } from '../../helpers/error.helper';
import { create } from '@bufbuild/protobuf';

@Component({
    selector: 'app-transaction-list',
    templateUrl: 'transactions-list.component.html',
    imports: [OverlayModule, FormsModule, ToastModule, TableModule, InputIcon, IconField, DatePipe, Button, MultiSelectModule, SelectModule, CommonModule]
})
export class TransactionsListComponent implements OnInit {
    private transactionsService;
    public loading = false;
    public transactions: Transaction[] = [];
    public accountsService;

    public filters: { [s: string]: FilterMetadata } = {};
    public accountMap: { [id: number]: Account } = {};

    private currentAccountId: number | undefined;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        public router: Router,
        private activeRoute: ActivatedRoute
    ) {
        this.transactionsService = createClient(TransactionsService, this.transport);
        this.accountsService = createClient(AccountsService, this.transport);
        this.currentAccountId = parseInt(this.activeRoute.snapshot.params['accountId']) ?? undefined;
    }

    async ngOnInit() {
        try {
            let resp = await this.accountsService.listAccounts({});
            for (let account of resp.accounts) {
                this.accountMap[account.account!.id] = account.account!;
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
            skip: event.first ?? 0
        });

        if (this.currentAccountId) {
            req.anyAccountIds = req.anyAccountIds ?? [];

            req.anyAccountIds.push(this.currentAccountId);
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

    getAccountColorClass(accID: number | undefined): string[] {
        let result = ['text-wrap', 'break-all'];

        if (this.currentAccountId && this.currentAccountId === accID) {
            result.push("text-purple-500");
        }

        return result
    }

    formatAmounts(transaction: Transaction): string {
        let val = '';

        let hasSource = false;

        if (transaction.sourceAmount) {
            val += `${transaction.sourceAmount} ${transaction.sourceCurrency ?? ''}`;
            hasSource = true;
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
