import { ChangeDetectionStrategy, Component, ElementRef, Inject, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { OverlayModule } from 'primeng/overlay';
import { FormsModule } from '@angular/forms';
import { InputText } from 'primeng/inputtext';
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
import { Router } from '@angular/router';
import { TransactionsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { Transaction } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { TimestampHelper } from '../../helpers/timestamp.helper';
import { AccountsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { AccountTypeEnum } from '../../services/enum.service';
import { Account } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { ErrorHelper } from '../../helpers/error.helper';

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

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        public router: Router
    ) {
        this.transactionsService = createClient(TransactionsService, this.transport);
        this.accountsService = createClient(AccountsService, this.transport);
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

    getAccountName(accountId: number| undefined): string {
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
        console.log(event)
        let resp = await this.transactionsService.listTransactions({
            limit: event.rows ?? 50,
            skip: event.first ?? 0
        })

        this.transactions = resp.transactions
    }

    formatAmounts(transaction: Transaction): string {
        let val = ""

        if (transaction.sourceAmount) {
            val += `${transaction.sourceAmount} ${transaction.sourceCurrency ?? ''}`;
        }

        if (transaction.destinationAmount) {
            val += ` (${transaction.destinationAmount} ${transaction.destinationCurrency ?? ''})`;
        }

        return val.trim();
    }

    protected readonly TimestampHelper = TimestampHelper;
}
