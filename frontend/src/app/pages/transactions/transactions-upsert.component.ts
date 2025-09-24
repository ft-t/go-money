import { Component, Inject, OnInit } from '@angular/core';
import { FluidModule } from 'primeng/fluid';
import { InputTextModule } from 'primeng/inputtext';
import { AbstractControl, FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { Transaction, TransactionSchema, TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { create } from '@bufbuild/protobuf';
import { AccountTypeEnum, EnumService } from '../../services/enum.service';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { ErrorHelper } from '../../helpers/error.helper';
import { AccountsService, ListAccountsResponse_AccountItem } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { FilterMetadata, MessageService } from 'primeng/api';
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
    GetApplicableAccountsResponse,
    GetApplicableAccountsResponse_ApplicableRecord, GetTitleSuggestionsRequestSchema,
    IncomeSchema,
    ListTransactionsRequestSchema,
    TransactionsService,
    TransferBetweenAccountsSchema,
    UpdateTransactionRequestSchema
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { CurrencyService, ExchangeRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/currency/v1/currency_pb';
import { Currency } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/currency_pb';
import { InputGroupModule } from 'primeng/inputgroup';
import { InputGroupAddonModule } from 'primeng/inputgroupaddon';
import { InputNumberModule } from 'primeng/inputnumber';
import { TagsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/tags/v1/tags_pb';
import { Tag, TagSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/tag_pb';
import { TimestampSchema } from '@bufbuild/protobuf/wkt';
import { SelectButtonModule } from 'primeng/selectbutton';
import { ChipModule } from 'primeng/chip';
import { ActivatedRoute, Router } from '@angular/router';
import { TimestampHelper } from '../../helpers/timestamp.helper';
import { CategoriesService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/categories/v1/categories_pb';
import { Category } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/category_pb';
import { SelectChangeEvent, SelectModule } from 'primeng/select';
import { Checkbox } from 'primeng/checkbox';
import { Message } from 'primeng/message';
import { greaterThanZeroValidator } from '../../validators/greaterthenzero';
import { AutoComplete, AutoCompleteCompleteEvent } from 'primeng/autocomplete';

type possibleDestination = 'source' | 'destination' | 'fx';

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
        Checkbox,
        Message,
        AutoComplete
    ]
})
export class TransactionUpsertComponent implements OnInit {
    public isEdit: boolean = false;

    // public transaction: Transaction;
    public transactionTypes: AccountTypeEnum[];
    public currencies: Currency[] = [];
    public tags: Tag[] = [];
    public categories: Category[] = [];

    public skipRules: false = false;
    public accounts: { [s: number]: GetApplicableAccountsResponse_ApplicableRecord } = {};
    public allAccounts: { [s: number]: Account } = {};

    private transactionService;

    private currencyService;
    private tagsService;
    private categoriesService;
    public maxSelectedLabels = 1;

    public form: FormGroup;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        private route: ActivatedRoute,
        private router: Router
    ) {
        this.transactionTypes = EnumService.getBaseTransactionTypes();
        this.transactionService = createClient(TransactionsService, this.transport);
        this.currencyService = createClient(CurrencyService, this.transport);
        this.tagsService = createClient(TagsService, this.transport);
        this.categoriesService = createClient(CategoriesService, this.transport);

        let targetType = TransactionType.EXPENSE;
        this.route.queryParams.subscribe(async (data) => {
            if (data['type']) {
                targetType = +(data['type'] as TransactionType);
            }
        });

        this.route.params.subscribe(async (params) => {
            if (params['id']) {
                await this.editTransaction(+params['id']);
            }
        });

        this.form = this.buildForm(
            create(TransactionSchema, {
                type: targetType
            })
        );
    }

    async ngOnInit() {
        await Promise.all([this.fetchAccounts(), this.fetchCurrencies(), this.fetchTags(), this.fetchCategories()]);
    }

    titleAutoComplete: string[] = [];

    async search(event: AutoCompleteCompleteEvent) {
        try {
            let resp = await this.transactionService.getTitleSuggestions(create(GetTitleSuggestionsRequestSchema, {
                limit: 50,
                query: event.query
            }))

            this.titleAutoComplete = resp.titles;

            if(!this.titleAutoComplete || this.titleAutoComplete.length == 0)
                this.titleAutoComplete = [event.query];

        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    buildForm(tx: Transaction) {
        const form = new FormGroup({
            id: new FormControl(tx.id, { nonNullable: false }),
            sourceAmount: new FormControl(this.toPositiveNumber(tx.sourceAmount), greaterThanZeroValidator()),
            sourceCurrency: new FormControl(tx.sourceCurrency, Validators.required),
            sourceAccountId: new FormControl(tx.sourceAccountId, Validators.min(1)),
            destinationAmount: new FormControl(this.toPositiveNumber(tx.destinationAmount), greaterThanZeroValidator()),
            destinationCurrency: new FormControl(tx.destinationCurrency, Validators.required),
            destinationAccountId: new FormControl(tx.destinationAccountId, Validators.min(1)),
            notes: new FormControl(tx.notes, { nonNullable: false }),
            title: new FormControl(tx.title, Validators.required),
            categoryId: new FormControl(tx.categoryId, { nonNullable: false }),
            transactionDate: new FormControl(tx.transactionDate != null ? TimestampHelper.timestampToDate(tx.transactionDate!) : new Date(), Validators.required),
            type: new FormControl(tx.type, Validators.required),
            tagIds: new FormControl(tx.tagIds || [], { nonNullable: false }),
            skipRules: new FormControl(this.skipRules, { nonNullable: false }),
            fxSourceAmount: new FormControl(this.toPositiveNumber(tx.fxSourceAmount), { nonNullable: false }),
            fxSourceCurrency: new FormControl(tx.fxSourceCurrency, { nonNullable: false })
        });

        form.get('destinationAccountId')!.valueChanges.subscribe(async (newVal) => {
            let curr = form.get('destinationCurrency');
            curr?.setValue('', { emitEvent: false });

            let account = this.getAccountById(newVal!);
            if (!account) {
                return;
            }

            curr?.setValue(account.currency, { emitEvent: false });
        });

        form.get('sourceAccountId')!.valueChanges.subscribe(async (newVal) => {
            let curr = form.get('sourceCurrency');
            curr?.setValue('', { emitEvent: false });

            let account = this.getAccountById(newVal!);
            if (!account) {
                return;
            }

            curr?.setValue(account.currency, { emitEvent: false });
        });

        // handle amounts
        form.get('destinationAmount')!.valueChanges.subscribe((newVal) => {
            if (form.get('sourceCurrency')!.value != form.get('destinationCurrency')!.value) return;

            form.get('sourceAmount')!.setValue(newVal, { emitEvent: false });
        });

        form.get('sourceAmount')!.valueChanges.subscribe((newVal) => {
            if (form.get('sourceCurrency')!.value != form.get('destinationCurrency')!.value) return;

            form.get('destinationAmount')!.setValue(newVal, { emitEvent: false });
        });

        form.get('type')!.valueChanges.subscribe((newVal) => {
            if (newVal != TransactionType.EXPENSE) {
                form.get('fxSourceAmount')?.setValue('');
                form.get('fxSourceCurrency')!.setValue('');
            }

            let sourceAccountId = form.get('sourceAccountId')!.value;
            let destinationAccountId = form.get('destinationAccountId')!.value;

            if (!this.isAccountApplicable(newVal!, true, sourceAccountId!)) {
                form.get('sourceAccountId')!.setValue(0);
            }

            if (!this.isAccountApplicable(newVal!, false, destinationAccountId!)) {
                form.get('destinationAccountId')!.setValue(0);
            }
        });

        return form;
    }

    get destinationAccountId() {
        return this.form.get('destinationAccountId')!;
    }

    get transactionDateFm() {
        return this.form.get('transactionDate')!;
    }

    get destinationAmount() {
        return this.form.get('destinationAmount')!;
    }

    get sourceAmount() {
        return this.form.get('sourceAmount')!;
    }

    get title() {
        return this.form.get('title')!;
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

            let tx = resp.transactions[0];

            if (tx.sourceAmount) tx.sourceAmount = this.toPositiveNumber(tx.sourceAmount)!;

            if (tx.destinationAmount) tx.destinationAmount = this.toPositiveNumber(tx.destinationAmount)!;

            this.form = this.buildForm(tx);
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    getTitle() {
        if (this.isEdit) {
            return `Edit Transaction (#${this.form.get('id')?.value})`;
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
            let resp = await this.transactionService.getApplicableAccounts({});
            for (let applicable of resp.applicableRecords) {
                this.accounts[applicable.transactionType] = applicable;

                for (let source of applicable.sourceAccounts || []) {
                    this.allAccounts[source.id] = source;
                }
                for (let destination of applicable.destinationAccounts || []) {
                    this.allAccounts[destination.id] = destination;
                }
            }
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    isAccountApplicable(type: TransactionType, isSource: boolean, id: number): boolean {
        let applicable = this.getApplicableAccounts(type, isSource);
        if (!applicable) {
            return false;
        }

        return applicable.some((a) => a.id == id);
    }

    getApplicableAccounts(type: TransactionType, isSource: boolean): Account[] {
        const applicable = this.accounts[type];

        if (!applicable) {
            return [];
        }

        if (isSource) {
            return applicable.sourceAccounts || [];
        } else {
            return applicable.destinationAccounts || [];
        }
    }

    // onTransactionTypeChange() {
    //     if (!this.isDestinationAccountActive()) {
    //         this.transaction.destinationAccountId = 0;
    //         this.transaction.destinationCurrency = '';
    //         this.transaction.destinationAmount = '';
    //     }
    //
    //     if (!this.isForeignCurrencyActive()) {
    //         this.transaction.destinationCurrency = '';
    //         this.transaction.destinationAmount = '';
    //     }
    // }

    getAccountById(id: number | undefined): Account | null {
        if (!id) return null;

        return this.allAccounts[id] || null;
    }

    canConvert(): boolean {
        let source = this.form.get('sourceCurrency')!.value;
        let dest = this.form.get('destinationCurrency')!.value;

        if (!source || !dest) {
            return false;
        }

        return source != dest;
    }

    log() {
        console.log(this.form);
    }

    async convertAmount(amount: number, currency: string, setTo: possibleDestination) {
        try {
            let destCurrency = '';
            let destValueSetter: AbstractControl | null;
            switch (setTo) {
                case 'source':
                    destCurrency = this.form.get('sourceCurrency')!.value;
                    destValueSetter = this.form.get('sourceAmount');
                    break;
                case 'destination':
                    destCurrency = this.form.get('destinationCurrency')!.value;
                    destValueSetter = this.form.get('destinationAmount');
                    break;
                case 'fx':
                    destCurrency = this.form.get('destinationCurrency')!.value;
                    destValueSetter = this.form.get('sourceAmount');
                    break;
            }

            let converted = await this.currencyService.exchange(
                create(ExchangeRequestSchema, {
                    amount: amount.toString(),
                    fromCurrency: currency,
                    toCurrency: destCurrency
                })
            );

            destValueSetter?.setValue(converted.amount);
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    removeTag(tag: number) {
        this.form.get('tagIds')!.setValue(this.form.get('tagIds')!.value.filter((t: number) => t != tag));
    }

    tagById(id: number | undefined): Tag | null {
        if (!id) return null;

        for (let tag of this.tags) {
            if (tag.id == id) return tag;
        }

        return null;
    }

    isForeignCurrencyActive(): boolean {
        if (this.form.get('type')?.value == TransactionType.EXPENSE) return true;

        return false;
    }

    buildTransactionRequest(): CreateTransactionRequest {
        let req = create(CreateTransactionRequestSchema, {
            notes: this.form.get('notes')!.value,
            extra: {}, // todo
            tagIds: this.form.get('tagIds')!.value || [],
            transactionDate: create(TimestampSchema, {
                seconds: BigInt(Math.floor(this.form.get('transactionDate')!.value.getTime() / 1000)),
                nanos: (this.form.get('transactionDate')!.value.getMilliseconds() % 1000) * 1_000_000
            }),
            title: this.form.get('title')!.value,
            categoryId: this.form.get('categoryId')!.value,
            skipRules: this.form.get('skipRules')!.value
        });

        let destinationAccountId = this.form.get('destinationAccountId')!.value;
        let destinationCurrency = this.form.get('destinationCurrency')!.value;
        let destinationAmount = this.form.get('destinationAmount')!.value;

        let sourceAccountId = this.form.get('sourceAccountId')!.value;
        let sourceCurrency = this.form.get('sourceCurrency')!.value;
        let sourceAmount = this.form.get('sourceAmount')!.value;

        let fxSourceAmount = this.form.get('fxSourceAmount')!.value;
        let fxSourceCurrency = this.form.get('fxSourceCurrency')!.value;

        switch (this.form.get('type')!.value) {
            case TransactionType.INCOME:
                req.transaction.value = create(IncomeSchema, {
                    destinationAccountId: destinationAccountId,
                    destinationAmount: this.toPositiveNumber(destinationAmount),
                    destinationCurrency: destinationCurrency,

                    sourceAccountId: sourceAccountId,
                    sourceAmount: this.toNegativeNumber(sourceAmount),
                    sourceCurrency: sourceCurrency
                });
                req.transaction.case = 'income';
                break;
            case TransactionType.EXPENSE:
                req.transaction.value = create(ExpenseSchema, {
                    destinationAccountId: destinationAccountId,
                    destinationAmount: this.toPositiveNumber(destinationAmount),
                    destinationCurrency: destinationCurrency,

                    sourceAccountId: sourceAccountId,
                    sourceAmount: this.toNegativeNumber(sourceAmount),
                    sourceCurrency: sourceCurrency,

                    fxSourceCurrency: fxSourceCurrency,
                    fxSourceAmount: this.toNegativeNumber(fxSourceAmount)
                });
                req.transaction.case = 'expense';
                break;
            case TransactionType.TRANSFER_BETWEEN_ACCOUNTS:
                req.transaction.value = create(TransferBetweenAccountsSchema, {
                    destinationAccountId: destinationAccountId,
                    destinationAmount: this.toPositiveNumber(destinationAmount),
                    destinationCurrency: destinationCurrency,

                    sourceAccountId: sourceAccountId,
                    sourceAmount: this.toNegativeNumber(sourceAmount),
                    sourceCurrency: sourceCurrency
                });
                req.transaction.case = 'transferBetweenAccounts';
                break;
        }

        return req;
    }

    async submit() {
        this.form!.markAllAsTouched();

        if (!this.form!.valid) {
            console.log(this.form);
            return;
        }

        if (this.isEdit) {
            await this.update();
        } else {
            await this.create();
        }
    }

    get sourceAccountId() {
        return this.form.get('sourceAccountId')!;
    }

    async create() {
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
        try {
            let response = await this.transactionService.updateTransaction(
                create(UpdateTransactionRequestSchema, {
                    transaction: this.buildTransactionRequest(),
                    id: this.form.get('id')?.value
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
        await this.editTransaction(Number(this.form.get('id')?.value));
    }
}
