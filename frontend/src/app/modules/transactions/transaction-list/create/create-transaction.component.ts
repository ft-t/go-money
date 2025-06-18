import { Component, Inject, OnInit } from '@angular/core';
import { DropdownChangeEvent, DropdownModule } from 'primeng/dropdown';
import { Fluid } from 'primeng/fluid';
import { InputText } from 'primeng/inputtext';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import {
    Transaction,
    TransactionSchema,
    TransactionType
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { create } from '@bufbuild/protobuf';
import { AccountTypeEnum, EnumService } from '../../../../services/enum.service';
import { TRANSPORT_TOKEN } from '../../../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { ErrorHelper } from '../../../../helpers/error.helper';
import {
    AccountsService,
    ListAccountsResponse_AccountItem
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { MessageService } from 'primeng/api';
import { Toast } from 'primeng/toast';
import { DatePicker } from 'primeng/datepicker';
import { IftaLabel } from 'primeng/iftalabel';
import { Account } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { NgIf } from '@angular/common';
import { Textarea } from 'primeng/textarea';
import { Button } from 'primeng/button';
import { MultiSelect } from 'primeng/multiselect';
import {
    CreateTransactionRequestSchema,
    DepositSchema,
    TransactionsService, WithdrawalSchema
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { CurrencyService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/currency/v1/currency_pb';
import { Currency } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/currency_pb';

@Component({
    selector: 'transaction-upsert',
    templateUrl: 'create-transaction.component.html',
    imports: [DropdownModule, Fluid, InputText, ReactiveFormsModule, FormsModule, Toast, DatePicker, IftaLabel, NgIf, Textarea, Button, MultiSelect]
})
export class TransactionUpsertComponent implements OnInit {
    public isEdit: boolean = false;

    public transaction: Transaction;
    public transactionTypes: AccountTypeEnum[];
    public labels: AccountTypeEnum[];
    public currencies: Currency[] = [];

    private accountService;
    public accounts: ListAccountsResponse_AccountItem[] = [];

    private transactionService;

    private currencyService;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService
    ) {
        this.transaction = create(TransactionSchema, {});
        this.transactionTypes = EnumService.getBaseTransactionTypes();
        this.accountService = createClient(AccountsService, this.transport);
        this.transactionService = createClient(TransactionsService, this.transport);
        this.currencyService = createClient(CurrencyService, this.transport);

        this.labels = [
            {
                name: 'tag1',
                value: 1
            },
            {
                name: 'tag2',
                value: 2
            }
        ];
    }

    async ngOnInit() {
        await this.fetchAccounts();
        await this.fetchCurrencies();
    }

    async fetchCurrencies() {
        try {
            let resp = await         this.currencyService.getCurrencies({})
            this.currencies = resp.currencies || [];
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    async fetchAccounts() {
        try {
            let resp = await this.accountService.listAccounts({});
            this.accounts = resp.accounts || [];
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    isSourceAccountActive(): boolean {
        if (this.transaction.type == TransactionType.WITHDRAWAL) return true;

        if (this.transaction.type == TransactionType.TRANSFER_BETWEEN_ACCOUNTS) return true;

        return false;
    }

    onTransactionTypeChange(event: DropdownChangeEvent) {
        if (!this.isDestinationAccountActive()) {
            this.transaction.destinationAccountId = 0;
            this.transaction.destinationCurrency = '';
            this.transaction.destinationAmount = '';
        }

        if (!this.isSourceAccountActive()) {
            this.transaction.sourceAccountId = 0;
            this.transaction.sourceCurrency = '';
            this.transaction.sourceAmount = '';
        }
    }

    onSourceAccountChange(event: DropdownChangeEvent) {
        this.transaction.sourceCurrency = this.accountById(this.transaction.sourceAccountId)?.currency ?? '';
    }

    onDestinationAccountChange(event: DropdownChangeEvent) {
        this.transaction.destinationCurrency = this.accountById(this.transaction.destinationAccountId)?.currency ?? '';
    }

    accountById(id: number | undefined): Account | null {
        if (!id) return null;

        for (let account of this.accounts) {
            if (account.account?.id == id) return account.account;
        }

        return null;
    }

    isDestinationAccountActive(): boolean {
        if (this.transaction.type == TransactionType.DEPOSIT) return true;

        if (this.transaction.type == TransactionType.TRANSFER_BETWEEN_ACCOUNTS) return true;

        return false;
    }

    async create() {
        let req = create(CreateTransactionRequestSchema, {
            notes: this.transaction.notes,
            extra: {}, // todo
            labelIds: this.transaction.labelIds,
            transactionDate: this.transaction.transactionDate,
            title: this.transaction.title,
        })

        switch (this.transaction.type) {
            case TransactionType.DEPOSIT:
                req.transaction.value = create(DepositSchema, {
                    destinationAccountId: this.transaction.destinationAccountId,
                    destinationAmount: this.transaction.destinationAmount,
                    destinationCurrency: this.transaction.destinationCurrency
                });
                req.transaction.case = "deposit"
                break;
            case TransactionType.WITHDRAWAL:
                req.transaction.value = create(WithdrawalSchema, {
                    sourceAmount:  this.transaction.sourceAmount,
                    sourceCurrency: this.transaction.sourceCurrency,
                    sourceAccountId: this.transaction.sourceAccountId
                });
                req.transaction.case = "withdrawal"
                break;
        }

        try {
            await this.transactionService.createTransaction(req)
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }
    async update() {}

    protected readonly TransactionType = TransactionType;
}
