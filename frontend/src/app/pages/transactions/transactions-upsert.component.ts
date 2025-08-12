import { Component, Inject, OnInit } from '@angular/core';
import { FluidModule } from 'primeng/fluid';
import { InputTextModule } from 'primeng/inputtext';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import {
    Transaction,
    TransactionSchema,
    TransactionType
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { create } from '@bufbuild/protobuf';
import { AccountTypeEnum, EnumService } from '../../services/enum.service';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { ErrorHelper } from '../../helpers/error.helper';
import {
    AccountsService,
    ListAccountsResponse_AccountItem
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { MessageService } from 'primeng/api';
import { ToastModule } from 'primeng/toast';
import { DatePickerModule } from 'primeng/datepicker';
import { Account } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { NgClass, NgIf } from '@angular/common';
import { TextareaModule } from 'primeng/textarea';
import { ButtonModule } from 'primeng/button';
import { MultiSelectModule } from 'primeng/multiselect';
import {
    CreateTransactionRequest,
    CreateTransactionRequestSchema,
    ExpenseSchema,
    IncomeSchema,
    ListTransactionsRequestSchema,
    TransactionsService,
    TransferBetweenAccountsSchema,
    UpdateTransactionRequestSchema,
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import {
    CurrencyService,
    ExchangeRequestSchema
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/currency/v1/currency_pb';
import { Currency } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/currency_pb';
import { InputGroupModule } from 'primeng/inputgroup';
import { InputGroupAddonModule } from 'primeng/inputgroupaddon';
import { InputNumberModule } from 'primeng/inputnumber';
import { TagsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/tags/v1/tags_pb';
import { Tag } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/tag_pb';
import { TimestampSchema } from '@bufbuild/protobuf/wkt';
import { SelectButtonModule } from 'primeng/selectbutton';
import { ChipModule } from 'primeng/chip';
import { ActivatedRoute, Router } from '@angular/router';
import { TimestampHelper } from '../../helpers/timestamp.helper';
import { CategoriesService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/categories/v1/categories_pb';
import { Category } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/category_pb';
import { SelectChangeEvent, SelectModule } from 'primeng/select';
import { Checkbox } from 'primeng/checkbox';

@Component({
    selector: 'transaction-upsert',
    templateUrl: 'transactions-upsert.component.html',
    imports: [
        SelectModule,
        SelectButtonModule,
        FluidModule,
        InputTextModule,
        ReactiveFormsModule,
        FormsModule,
        ToastModule,
        DatePickerModule,
        NgIf,
        TextareaModule,
        ButtonModule,
        MultiSelectModule,
        InputGroupModule,
        InputGroupAddonModule,
        InputNumberModule,
        SelectButtonModule,
        ChipModule,
        NgClass,
        Checkbox
    ]
})
export class TransactionUpsertComponent implements OnInit {
    public isEdit: boolean = false;

    public transaction: Transaction;
    public transactionTypes: AccountTypeEnum[];
    public currencies: Currency[] = [];
    public tags: Tag[] = [];
    public categories: Category[] = [];

    public skipRules: false = false;
    private accountService;
    public accounts: ListAccountsResponse_AccountItem[] = [];

    private transactionService;

    private currencyService;
    private tagsService;
    private categoriesService;
    public transactionDate: Date = new Date();

    public showValidation = false;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        private route: ActivatedRoute,
        private router: Router
    ) {
        this.transaction = create(TransactionSchema, {
            destinationAmount: undefined,
            sourceAmount: undefined,
            type: TransactionType.EXPENSE
        });

        this.transactionTypes = EnumService.getBaseTransactionTypes();
        this.accountService = createClient(AccountsService, this.transport);
        this.transactionService = createClient(TransactionsService, this.transport);
        this.currencyService = createClient(CurrencyService, this.transport);
        this.tagsService = createClient(TagsService, this.transport);
        this.categoriesService = createClient(CategoriesService, this.transport);

        this.route.queryParams.subscribe(async (data) => {
            if (data['type']) {
                this.transaction.type = +(data['type'] as TransactionType);
            }
        });

        this.route.params.subscribe(async (params) => {
            if (params['id']) {
                await this.editTransaction(+params['id']);
            }
        });
    }

    async ngOnInit() {
        await Promise.all([this.fetchAccounts(), this.fetchCurrencies(), this.fetchTags(), this.fetchCategories()]);
    }

    buildForm() {
        // this.form = new FormGroup({
        //
        // });
    }

    async editTransaction(id: number) {
        this.isEdit = true;

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

            if (this.transaction.sourceAmount) this.transaction.sourceAmount = this.toPositiveNumber(this.transaction.sourceAmount)!;

            if (this.transaction.destinationAmount) this.transaction.destinationAmount = this.toPositiveNumber(this.transaction.destinationAmount)!;

            this.transactionDate = TimestampHelper.timestampToDate(this.transaction.transactionDate!);
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    getTitle() {
        if (this.isEdit) {
            return `Editing Transaction (#${this.transaction.id})`;
        }

        return `New Transaction`;
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

    async fetchCategories() {
        try {
            let resp = await this.categoriesService.listCategories({});
            this.categories = resp.categories || [];
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
        if (this.transaction.type == TransactionType.EXPENSE) return true;

        if (this.transaction.type == TransactionType.TRANSFER_BETWEEN_ACCOUNTS) return true;

        return false;
    }

    onTransactionTypeChange() {
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

        if (!this.isForeignCurrencyActive()) {
            this.transaction.destinationCurrency = '';
            this.transaction.destinationAmount = '';
        }
    }

    onSourceAccountChange(_: SelectChangeEvent) {
        this.transaction.sourceCurrency = this.accountById(this.transaction.sourceAccountId)?.currency ?? '';
    }

    onDestinationAccountChange(_: SelectChangeEvent) {
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

    log() {
        console.log(this.transaction);
    }

    async convertAmount() {
        if (!this.canConvertAmount()) return;

        try {
            let converted = await this.currencyService.exchange(
                create(ExchangeRequestSchema, {
                    amount: this.transaction!.sourceAmount!.toString(),
                    fromCurrency: this.transaction.sourceCurrency,
                    toCurrency: this.transaction.destinationCurrency
                })
            );

            this.transaction.destinationAmount = converted.amount;
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    maxSelectedLabels = 1;

    removeTag(tag: number) {
        this.transaction.tagIds = this.transaction.tagIds.filter((t) => t != tag);
    }

    tagById(id: number | undefined): Tag | null {
        if (!id) return null;

        for (let tag of this.tags) {
            if (tag.id == id) return tag;
        }

        return null;
    }

    isDestinationAccountActive(): boolean {
        if (this.transaction.type == TransactionType.INCOME) return true;

        if (this.transaction.type == TransactionType.TRANSFER_BETWEEN_ACCOUNTS) return true;

        return false;
    }

    isForeignCurrencyActive(): boolean {
        if (this.transaction.type == TransactionType.EXPENSE) return true;

        return false;
    }

    buildTransactionRequest(): CreateTransactionRequest {
        let req = create(CreateTransactionRequestSchema, {
            notes: this.transaction.notes,
            extra: {}, // todo
            tagIds: this.transaction.tagIds,
            transactionDate: create(TimestampSchema, {
                seconds: BigInt(Math.floor(this.transactionDate.getTime() / 1000)),
                nanos: (this.transactionDate.getMilliseconds() % 1000) * 1_000_000
            }),
            title: this.transaction.title,
            categoryId: this.transaction.categoryId,
            skipRules: this.skipRules
        });

        switch (this.transaction.type) {
            case TransactionType.INCOME:
                req.transaction.value = create(IncomeSchema, {
                    destinationAccountId: this.transaction.destinationAccountId,
                    destinationAmount: this.toPositiveNumber(this.transaction.destinationAmount),
                    destinationCurrency: this.transaction.destinationCurrency
                });
                req.transaction.case = 'income';
                break;
            case TransactionType.EXPENSE:
                req.transaction.value = create(ExpenseSchema, {
                    sourceAmount: this.toNegativeNumber(this.transaction.sourceAmount),
                    sourceCurrency: this.transaction.sourceCurrency,
                    sourceAccountId: this.transaction.sourceAccountId,
                    fxSourceAmount: this.toNegativeNumber(this.transaction.destinationAmount), // todo
                    fxSourceCurrency: this.transaction.destinationCurrency // todo
                });
                req.transaction.case = 'expense';
                break;
            case TransactionType.TRANSFER_BETWEEN_ACCOUNTS:
                req.transaction.value = create(TransferBetweenAccountsSchema, {
                    sourceAccountId: this.transaction.sourceAccountId,
                    destinationAccountId: this.transaction.destinationAccountId,
                    sourceAmount: this.toNegativeNumber(this.transaction.sourceAmount),
                    sourceCurrency: this.transaction.sourceCurrency,
                    destinationAmount: this.toPositiveNumber(this.transaction.destinationAmount),
                    destinationCurrency: this.transaction.destinationCurrency
                });
                req.transaction.case = 'transferBetweenAccounts';
                break;
        }

        return req;
    }

    async create() {
        this.showValidation = true;
        if (!this.isFormValid()) {
            this.messageService.add({ severity: 'error', detail: 'Please fill all required fields.' });
            return;
        }

        try {
            await this.transactionService.createTransaction(this.buildTransactionRequest());

            this.messageService.add({ severity: 'info', detail: 'Transaction created successfully.' });

            // todo transaction details page
            await this.router.navigate(['/transactions']);
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    async update() {
        this.showValidation = true;
        if (!this.isFormValid()) {
            this.messageService.add({ severity: 'error', detail: 'Please fill all required fields.' });
            return;
        }

        try {
            let response = await this.transactionService.updateTransaction(
                create(UpdateTransactionRequestSchema, {
                    transaction: this.buildTransactionRequest(),
                    id: this.transaction.id
                })
            );

            this.messageService.add({ severity: 'info', detail: 'Transaction created successfully.' });

            // todo transaction details page
            await this.router.navigate(['/transactions']);

            // todo transaction details page
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    isFormValid(): boolean {
        if (!this.transaction.title) return false;
        if (!this.transactionDate) return false;
        if (this.isSourceAccountActive() && !this.transaction.sourceAccountId) return false;
        if (this.isSourceAccountActive() && !this.transaction.sourceAmount) return false;
        if (this.isDestinationAccountActive() && !this.transaction.destinationAccountId) return false;
        if (this.isDestinationAccountActive() && !this.transaction.destinationAmount) return false;
        return true;
    }

    toNegativeNumber(value: string | undefined): string | undefined {
        if (!value) return value;

        let num = parseFloat(value);
        if (!num) return value;

        return (-Math.abs(num)).toString();
    }

    toPositiveNumber(value: string | undefined): string | undefined {
        if (!value) return value;

        let num = parseFloat(value);
        if (!num) return value;

        return Math.abs(num).toString();
    }

    async refresh() {
        await this.editTransaction(Number(this.transaction.id));
    }
}
