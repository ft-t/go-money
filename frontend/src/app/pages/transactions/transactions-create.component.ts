import { Component, Inject, OnInit } from '@angular/core';
import { DropdownChangeEvent, DropdownModule } from 'primeng/dropdown';
import { Fluid } from 'primeng/fluid';
import { InputText } from 'primeng/inputtext';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { Transaction, TransactionSchema, TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { create } from '@bufbuild/protobuf';
import { AccountTypeEnum, EnumService } from '../../services/enum.service';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { ErrorHelper } from '../../helpers/error.helper';
import { AccountsService, ListAccountsResponse_AccountItem } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
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
    ListTransactionsRequestSchema,
    TransactionsService,
    WithdrawalSchema
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import {
    CurrencyService,
    ExchangeRequestSchema
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/currency/v1/currency_pb';
import { Currency } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/currency_pb';
import { InputGroup } from 'primeng/inputgroup';
import { InputGroupAddon } from 'primeng/inputgroupaddon';
import { InputNumber } from 'primeng/inputnumber';
import { TagsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/tags/v1/tags_pb';
import { Tag } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/tag_pb';
import { TimestampSchema } from '@bufbuild/protobuf/wkt';
import { SelectButton, SelectButtonModule } from 'primeng/selectbutton';
import { Chip } from 'primeng/chip';
import { ActivatedRoute } from '@angular/router';

@Component({
    selector: 'transaction-upsert',
    templateUrl: 'transactions-create.component.html',
    imports: [SelectButtonModule, DropdownModule, Fluid, InputText, ReactiveFormsModule, FormsModule, Toast, DatePicker, NgIf, Textarea, Button, MultiSelect, InputGroup, InputGroupAddon, InputNumber, SelectButton, Chip]
})
export class TransactionUpsertComponent implements OnInit {
    public isEdit: boolean = false;

    public transaction: Transaction;
    public transactionTypes: AccountTypeEnum[];
    public currencies: Currency[] = [];
    public tags: Tag[] = [];

    private accountService;
    public accounts: ListAccountsResponse_AccountItem[] = [];

    private transactionService;

    private currencyService;
    private tagsService;
    public transactionDate: Date = new Date();

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        private route: ActivatedRoute,
    ) {
        this.transaction = create(TransactionSchema, {
            destinationAmount: undefined,
            sourceAmount: undefined,
            type: TransactionType.WITHDRAWAL
        });

        this.transactionTypes = EnumService.getBaseTransactionTypes();
        this.accountService = createClient(AccountsService, this.transport);
        this.transactionService = createClient(TransactionsService, this.transport);
        this.currencyService = createClient(CurrencyService, this.transport);
        this.tagsService = createClient(TagsService, this.transport);

        this.route.queryParams.subscribe(async (data) => {
            if(data['type']) {
                this.transaction.type = +(data['type'] as TransactionType);
            }

            if(data['id']) {
                await this.editTransaction(+data['id']);
            }
        })
    }

    async ngOnInit() {
        await Promise.all([this.fetchAccounts(), this.fetchCurrencies(), this.fetchTags()]);
    }

    async editTransaction(id: number) {
        this.transactionService.listTransactions(create(ListTransactionsRequestSchema, {

        }))
    }

    getTitle() {
        if (this.isEdit) {
            return 'Editing Transaction';
        }

        return 'New Transaction';
    }

    async fetchTags() {
        try {
            let resp = await this.tagsService.listTags({});
            for (let tag of resp.tags || []) {
                this.tags.push(tag.tag!);
            }
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    async fetchCurrencies() {
        try {
            let resp = await this.currencyService.getCurrencies({});
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

    onTransactionTypeChange() {
        if (!this.isDestinationAccountActive()) {
            this.transaction.destinationAccountId = undefined;
            this.transaction.destinationCurrency = undefined;
            this.transaction.destinationAmount = undefined;
        }

        if (!this.isSourceAccountActive()) {
            this.transaction.sourceAccountId = undefined;
            this.transaction.sourceCurrency = undefined;
            this.transaction.sourceAmount = undefined;
        }

        if (!this.isForeignCurrencyActive()) {
            this.transaction.destinationCurrency = undefined;
            this.transaction.destinationAmount = undefined;
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

    canConvertAmount(): boolean {
        if (!this.transaction.sourceAmount || !this.transaction.sourceCurrency || !this.transaction.destinationCurrency) {
            return false;
        }

        return true;
    }
    async convertAmount() {
        if(!this.canConvertAmount())
            return;

        try{
            let converted = await this.currencyService.exchange(create(ExchangeRequestSchema, {
                amount: this.transaction!.sourceAmount!.toString(),
                fromCurrency: this.transaction.sourceCurrency,
                toCurrency: this.transaction.destinationCurrency
            }))

            this.transaction.destinationAmount = converted.amount;
        }
        catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    maxSelectedLabels = 1

    removeTag(tag: number) {
        this.transaction.tagIds = this.transaction.tagIds.filter(t => t != tag);
    }

    tagById(id: number | undefined): Tag | null {
        if (!id) return null;

        for (let tag of this.tags) {
            if (tag.id == id) return tag;
        }

        return null;
    }

    isDestinationAccountActive(): boolean {
        if (this.transaction.type == TransactionType.DEPOSIT) return true;

        if (this.transaction.type == TransactionType.TRANSFER_BETWEEN_ACCOUNTS) return true;

        return false;
    }

    isForeignCurrencyActive(): boolean {
        if (this.transaction.type == TransactionType.WITHDRAWAL) return true;

        return false;
    }

    async create() {
        let req = create(CreateTransactionRequestSchema, {
            notes: this.transaction.notes,
            extra: {}, // todo
            tagIds: this.transaction.tagIds,
            transactionDate: create(TimestampSchema, {
                seconds: BigInt(Math.floor(this.transactionDate.getTime() / 1000)),
                nanos: (this.transactionDate.getMilliseconds() % 1000) * 1_000_000
            }),
            title: this.transaction.title
        });

        switch (this.transaction.type) {
            case TransactionType.DEPOSIT:
                req.transaction.value = create(DepositSchema, {
                    destinationAccountId: this.transaction.destinationAccountId,
                    destinationAmount: this.transaction.destinationAmount,
                    destinationCurrency: this.transaction.destinationCurrency
                });
                req.transaction.case = 'deposit';
                break;
            case TransactionType.WITHDRAWAL:
                req.transaction.value = create(WithdrawalSchema, {
                    sourceAmount: this.toNegativeNumber(this.transaction.sourceAmount),
                    sourceCurrency: this.transaction.sourceCurrency,
                    sourceAccountId: this.transaction.sourceAccountId,
                    foreignAmount: this.toNegativeNumber(this.transaction.destinationAmount),
                    foreignCurrency: this.transaction.destinationCurrency
                });
                req.transaction.case = 'withdrawal';

                break;
        }

        try {
            await this.transactionService.createTransaction(req);
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    async update() {}

    toNegativeNumber(value: string | undefined): string | undefined {
        if (!value) return value;

        let num = parseFloat(value);
        if (!num) return value;

        return (-Math.abs(num)).toString();
    }

    protected readonly TransactionType = TransactionType;
}
