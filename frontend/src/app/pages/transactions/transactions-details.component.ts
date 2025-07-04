import { Component, Inject, OnInit } from '@angular/core';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { MessageService } from 'primeng/api';
import { ActivatedRoute } from '@angular/router';
import { ListTransactionsRequestSchema, TransactionsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { ErrorHelper } from '../../helpers/error.helper';
import { Transaction, TransactionSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { create } from '@bufbuild/protobuf';
import { Fluid } from 'primeng/fluid';
import { Card } from 'primeng/card';
import { Tag } from 'primeng/tag';
import { NgForOf, NgIf } from '@angular/common';
import { Account } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import {
    AccountsService,
    ListAccountsResponse_AccountItem
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { TimestampHelper } from '../../helpers/timestamp.helper';

@Component({
    selector: 'app-transaction-details',
    imports: [Fluid, Card, Tag, NgForOf, NgIf],
    templateUrl: './transactions-details.component.html'
})
export class TransactionsDetailsComponent implements OnInit {
    public transaction: Transaction = create(TransactionSchema, {});
    private transactionService;
    public accounts: ListAccountsResponse_AccountItem[] = [];
    private accountService;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        private route: ActivatedRoute
    ) {
        this.transactionService = createClient(TransactionsService, this.transport);
        this.accountService = createClient(AccountsService, this.transport);

        this.route.params.subscribe(async (params) => {
            if (params['id']) {
                await this.fetchTransaction(+params['id']);
            }
        });
    }

    async ngOnInit() {
        await Promise.all([this.fetchAccounts()]);
    }

    async fetchAccounts() {
        try {
            let resp = await this.accountService.listAccounts({});
            this.accounts = resp.accounts || [];
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    async fetchTransaction(id: number) {
        try {
            let resp = await this.transactionService.listTransactions(
                create(ListTransactionsRequestSchema, {
                    ids: [id],
                    limit: 1
                })
            );

            if (!resp.transactions || resp.transactions.length == 0) {
                this.messageService.add({ severity: 'error', detail: 'Transaction not found.' });
                return;
            }

            this.transaction = resp.transactions[0];

            if (this.transaction.sourceAmount) this.transaction.sourceAmount = this.toPositiveNumber(this.transaction.sourceAmount);

            if (this.transaction.destinationAmount) this.transaction.destinationAmount = this.toPositiveNumber(this.transaction.destinationAmount);
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    accountById(id: number | undefined): Account | null {
        if (!id) return null;

        for (let account of this.accounts) {
            if (account.account?.id == id) return account.account;
        }

        return null;
    }

    toPositiveNumber(value: string | undefined): string | undefined {
        if (!value) return value;

        let num = parseFloat(value);
        if (!num) return value;

        return Math.abs(num).toString();
    }

    protected readonly TimestampHelper = TimestampHelper;
}
