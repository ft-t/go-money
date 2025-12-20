import { Component, Inject, OnDestroy, OnInit, QueryList, ViewChildren } from '@angular/core';
import { FluidModule } from 'primeng/fluid';
import { InputTextModule } from 'primeng/inputtext';
import { AbstractControl, FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { Transaction, TransactionSchema, TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { create } from '@bufbuild/protobuf';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { ErrorHelper } from '../../helpers/error.helper';
import { FilterMetadata, MessageService, ConfirmationService } from 'primeng/api';
import { ToastModule } from 'primeng/toast';
import { DatePickerModule } from 'primeng/datepicker';
import { NgClass, NgIf } from '@angular/common';
import { TextareaModule } from 'primeng/textarea';
import { ButtonModule } from 'primeng/button';
import { MultiSelectModule } from 'primeng/multiselect';
import {
    CreateTransactionRequest,
    CreateTransactionRequestSchema,
    DeleteTransactionsRequestSchema,
    ExpenseSchema,
    GetApplicableAccountsResponse,
    GetApplicableAccountsResponse_ApplicableRecord,
    GetTitleSuggestionsRequestSchema,
    IncomeSchema,
    ListTransactionsRequestSchema,
    TransactionsService,
    TransferBetweenAccountsSchema,
    UpdateTransactionRequestSchema
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { Account, AccountType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { CurrencyService, ExchangeRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/currency/v1/currency_pb';
import { InputGroupModule } from 'primeng/inputgroup';
import { InputGroupAddonModule } from 'primeng/inputgroupaddon';
import { InputNumberModule } from 'primeng/inputnumber';
import { TimestampSchema } from '@bufbuild/protobuf/wkt';
import { SelectButtonModule } from 'primeng/selectbutton';
import { ChipModule } from 'primeng/chip';
import { ActivatedRoute, Router } from '@angular/router';
import { SelectModule } from 'primeng/select';
import { greaterThanZeroValidator } from '../../validators/greaterthenzero';
import { ConfirmDialogModule } from 'primeng/confirmdialog';
import { TransactionEditorComponent } from '../../shared/components/transaction-editor/transaction-editor.component';
import { NumberHelper } from '../../helpers/number.helper';
import { AccountHelper } from '../../helpers/account.helper';
import { Tooltip } from 'primeng/tooltip';
import { Dialog } from 'primeng/dialog';
import { Message } from 'primeng/message';
import { Subject, takeUntil } from 'rxjs';
import { TimestampHelper } from '../../helpers/timestamp.helper';
import { SnippetSpotlightComponent } from '../../shared/components/snippet-spotlight/snippet-spotlight.component';

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
        ConfirmDialogModule,
        TransactionEditorComponent,
        Tooltip,
        Dialog,
        Message,
        SnippetSpotlightComponent
    ]
})
export class TransactionUpsertComponent implements OnInit, OnDestroy {
    private transactionService;
    private currencyService;
    private lastTxID = 0;
    public targetTransaction: Transaction[] = [];
    public showExpenseSplit = false;
    public expenseSplitForm: FormGroup | undefined = undefined;
    private currentSplitIndex = 0;
    public accounts: { [s: number]: GetApplicableAccountsResponse_ApplicableRecord } = {};
    public allAccounts: { [s: number]: Account } = {};
    public isReconciliationTransaction = false;
    public snippetSpotlightVisible = false;
    private destroy$ = new Subject<void>();

    @ViewChildren('editor') components: QueryList<TransactionEditorComponent> = new QueryList();

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        private confirmationService: ConfirmationService,
        private router: Router,
        private route: ActivatedRoute
    ) {
        this.transactionService = createClient(TransactionsService, this.transport);
        this.currencyService = createClient(CurrencyService, this.transport);

        let targetType = TransactionType.EXPENSE;
        this.route.queryParams.subscribe(async (data) => {
            if (data['type']) {
                targetType = +(data['type'] as TransactionType);
            }

            if (data['clone_from']) {
                await this.setTx(+data['clone_from'], 0);
                this.targetTransaction[0]!.id = BigInt(0);
            }
        });

        this.route.params.subscribe(async (params) => {
            if (params['id']) {
                await this.setTx(+params['id'], 0);
            }
        });

        this.targetTransaction[0] = create(TransactionSchema, {
            type: targetType
        });
    }

    async ngOnInit(): Promise<void> {
        await this.fetchAccounts();
    }

    async refresh() {
        await this.fetchAccounts();
    }

    showCommissionSplit(index: number) {
        this.currentSplitIndex = index;
        let editor = this.components.get(index);
        let result = editor?.getForm();

        const sourceAccountId = result?.get('sourceAccountId')?.value;
        const sourceCurrency = result?.get('sourceCurrency')?.value;
        const sourceAmount = parseFloat(result?.get('sourceAmount')?.value || '0');
        const transactionDate = result?.get('transactionDate')?.value;

        const defaultDestinationAccount = AccountHelper.getExpenseAccountByNameOrDefault(this.accounts, 'bank commissions');
        const destinationAccountId = defaultDestinationAccount?.id || 0;
        const destinationCurrency = defaultDestinationAccount?.currency || sourceCurrency;

        this.expenseSplitForm = new FormGroup({
            sourceAccountId: new FormControl(sourceAccountId),
            sourceAccountName: new FormControl(editor?.getSourceAccountName() || 'Unknown'),
            title: new FormControl('Commission Split', Validators.required),
            sourceCurrency: new FormControl(sourceCurrency),
            transactionDate: new FormControl(transactionDate),
            destinationAccountId: new FormControl(destinationAccountId, Validators.min(1)),
            destinationCurrency: new FormControl(destinationCurrency),
            amount: new FormControl(undefined, [Validators.required, greaterThanZeroValidator, Validators.max(Math.abs(sourceAmount))])
        });

        this.expenseSplitForm.get('sourceAccountName')!.disable();

        this.expenseSplitForm.get('destinationAccountId')!.valueChanges
            .pipe(takeUntil(this.destroy$))
            .subscribe(async (newVal) => {
                let curr = this.expenseSplitForm!.get('destinationCurrency');
                curr?.setValue('', { emitEvent: false });

                let account = AccountHelper.getAccountById(this.allAccounts, newVal!);
                if (!account) {
                    return;
                }

                curr?.setValue(account.currency, { emitEvent: false });
            });

        this.showExpenseSplit = true;
    }

    onHideExpenseSplit() {
        this.destroy$.next();
        this.showExpenseSplit = false;
        this.expenseSplitForm = undefined;
    }

    async saveExpenseSplit() {
        if (!this.expenseSplitForm) return;

        this.expenseSplitForm.markAllAsTouched();

        if (!this.expenseSplitForm.valid) {
            return;
        }

        const amount = this.expenseSplitForm.get('amount')!.value;
        const delta = NumberHelper.toPositiveNumber(amount?.toString());

        if (!delta) {
            return;
        }

        const editor = this.components.get(this.currentSplitIndex);
        let originalForm = editor!.getForm();

        const sourceCurrency = this.expenseSplitForm.get('sourceCurrency')!.value;
        const destinationCurrency = this.expenseSplitForm.get('destinationCurrency')!.value;
        let destinationAmount = delta;

        if (sourceCurrency !== destinationCurrency) {
            try {
                let converted = await this.currencyService.exchange(
                    create(ExchangeRequestSchema, {
                        amount: delta,
                        fromCurrency: sourceCurrency,
                        toCurrency: destinationCurrency
                    })
                );
                destinationAmount = converted.amount;
            } catch (e) {
                this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
                return;
            }
        }

        const newTransaction = create(TransactionSchema, {
            type: TransactionType.EXPENSE,
            sourceAccountId: this.expenseSplitForm.get('sourceAccountId')!.value,
            sourceCurrency: sourceCurrency,
            sourceAmount: delta,
            destinationAccountId: this.expenseSplitForm.get('destinationAccountId')!.value,
            destinationCurrency: destinationCurrency,
            destinationAmount: destinationAmount,
            title: this.expenseSplitForm.get('title')!.value,
            transactionDate: create(TimestampSchema, TimestampHelper.dateToTimestamp(originalForm!.get('transactionDate')!.value)),
        });

        console.log(newTransaction);


        this.targetTransaction.push(newTransaction);


        if (editor) {
            editor.adjustSourceAmount(parseFloat(NumberHelper.toNegativeNumber(delta)!));
        }

        this.messageService.add({
            severity: 'success',
            detail: 'Commission split added successfully'
        });

        this.onHideExpenseSplit();
    }

    getTitle(tx: Transaction): string {
        const id = tx.id;

        if (id) {
            return `Edit Transaction (#${id})`;
        }

        return `New Transaction`;
    }

    async setTx(id: number, index: number) {
        this.lastTxID = id;

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

        if (tx.type === TransactionType.ADJUSTMENT) {
            this.isReconciliationTransaction = true;
            this.targetTransaction[index] = tx;
            return;
        }

        if (tx.sourceAmount) tx.sourceAmount = NumberHelper.toPositiveNumber(tx.sourceAmount)!;

        if (tx.destinationAmount) tx.destinationAmount = NumberHelper.toPositiveNumber(tx.destinationAmount)!;

        this.targetTransaction[index] = tx;
    }

    // async confirmDelete() {
    //     this.confirmationService.confirm({
    //         message: 'Are you sure you want to delete this transaction? This action cannot be undone.',
    //         header: 'Confirm Delete',
    //         icon: 'pi pi-exclamation-triangle',
    //         acceptButtonStyleClass: 'p-button-danger',
    //         accept: async () => {
    //             await this.deleteTransaction();
    //         }
    //     });
    // }

    getTransactionCounts(): { create: number; update: number } {
        let create = 0;
        let update = 0;

        for (const tx of this.targetTransaction) {
            if (tx.id === BigInt(0)) {
                create++;
            } else {
                update++;
            }
        }

        return { create, update };
    }

    getSaveButtonLabel(): string {
        return 'Save All';
    }

    getSaveButtonDescription(): string {
        const counts = this.getTransactionCounts();
        const parts: string[] = [];

        if (counts.create > 0) {
            parts.push(`${counts.create} to create`);
        }
        if (counts.update > 0) {
            parts.push(`${counts.update} to update`);
        }

        return parts.length > 0 ? parts.join(', ') : '';
    }

    async saveAll() {
        try {
            for (const editor of this.components) {
                const form = editor.getForm();
                form.markAllAsTouched();

                if (!form.valid) {
                    this.messageService.add({
                        severity: 'error',
                        detail: 'Please fix validation errors in all transactions'
                    });
                    return;
                }
            }

            const counts = this.getTransactionCounts();
            let successCount = 0;

            for (let i = 0; i < this.targetTransaction.length; i++) {
                const tx = this.targetTransaction[i];
                const editor = this.components.get(i);

                if (!editor) continue;

                const request = editor.buildTransactionRequest();

                if (tx.id === BigInt(0)) {
                    // Create new transaction
                    await this.transactionService.createTransaction(request);
                    successCount++;
                } else {
                    // Update existing transaction
                    await this.transactionService.updateTransaction(
                        create(UpdateTransactionRequestSchema, {
                            transaction: request,
                            id: tx.id
                        })
                    );
                    successCount++;
                }
            }

            this.messageService.add({
                severity: 'success',
                detail: `Successfully saved ${successCount} transaction(s)`
            });

            await this.router.navigate(['/transactions']);
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    async addSplit() {
        console.log(this.components);
        this.targetTransaction.push(
            create(TransactionSchema, {
                type: this.targetTransaction[0]!.type
            })
        );
    }

    canDeleteSplit(index: number, tx: Transaction): boolean {
        return index !== 0 && tx.id === BigInt(0);
    }

    deleteSplit(index: number) {
        if (!this.canDeleteSplit(index, this.targetTransaction[index])) {
            return;
        }

        this.targetTransaction.splice(index, 1);
    }

    canDeleteTransaction(tx: Transaction): boolean {
        return tx.id !== BigInt(0);
    }

    confirmDelete(index: number) {
        const tx = this.targetTransaction[index];
        if (!this.canDeleteTransaction(tx)) {
            return;
        }

        this.confirmationService.confirm({
            message: 'Are you sure you want to delete this transaction? This action cannot be undone.',
            header: 'Confirm Delete',
            icon: 'pi pi-exclamation-triangle',
            acceptButtonStyleClass: 'p-button-danger',
            accept: async () => {
                await this.deleteTransaction(index);
            }
        });
    }

    async deleteTransaction(index: number) {
        try {
            const tx = this.targetTransaction[index];
            if (!tx || !tx.id) {
                this.messageService.add({ severity: 'error', detail: 'No transaction ID found' });
                return;
            }

            await this.transactionService.deleteTransactions(
                create(DeleteTransactionsRequestSchema, {
                    ids: [tx.id]
                })
            );

            this.messageService.add({ severity: 'success', detail: 'Transaction deleted successfully.' });
            await this.router.navigate(['/transactions']);
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    protected readonly BigInt = BigInt;

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

    getApplicableAccounts(type: TransactionType, isSource: boolean): Account[] {
        return AccountHelper.getApplicableAccounts(this.accounts, type, isSource);
    }

    getAccountTypeName(type: number): string {
        return AccountHelper.getAccountTypeName(type);
    }

    parseFloat(value: string): number {
        return AccountHelper.parseFloat(value);
    }

    toggleSnippetSpotlight(): void {
        this.snippetSpotlightVisible = !this.snippetSpotlightVisible;
    }

    getCurrentTransactionsAsSnippet(): Transaction[] {
        const result: Transaction[] = [];

        for (let i = 0; i < this.components.length; i++) {
            const editor = this.components.get(i);
            if (!editor) continue;

            const form = editor.getForm();

            result.push(create(TransactionSchema, {
                type: form.get('type')?.value ?? 0,
                title: form.get('title')?.value ?? '',
                notes: form.get('notes')?.value ?? '',
                sourceAccountId: form.get('sourceAccountId')?.value ?? 0,
                sourceCurrency: form.get('sourceCurrency')?.value ?? '',
                sourceAmount: form.get('sourceAmount')?.value?.toString() ?? '',
                destinationAccountId: form.get('destinationAccountId')?.value ?? 0,
                destinationCurrency: form.get('destinationCurrency')?.value ?? '',
                destinationAmount: form.get('destinationAmount')?.value?.toString() ?? '',
                categoryId: form.get('categoryId')?.value ?? 0,
                tagIds: form.get('tagIds')?.value ?? [],
                fxSourceAmount: form.get('fxSourceAmount')?.value?.toString() ?? '',
                fxSourceCurrency: form.get('fxSourceCurrency')?.value ?? '',
                internalReferenceNumbers: form.get('internalReferenceNumbers')?.value ?? []
            }));
        }

        return result;
    }

    applySnippetTransactions(transactions: Transaction[]): void {
        this.targetTransaction = transactions.map(tx =>
            create(TransactionSchema, {
                type: tx.type,
                title: tx.title,
                notes: tx.notes,
                sourceAccountId: tx.sourceAccountId,
                sourceCurrency: tx.sourceCurrency,
                sourceAmount: tx.sourceAmount,
                destinationAccountId: tx.destinationAccountId,
                destinationCurrency: tx.destinationCurrency,
                destinationAmount: tx.destinationAmount,
                categoryId: tx.categoryId,
                tagIds: tx.tagIds,
                transactionDate: create(TimestampSchema, TimestampHelper.dateToTimestamp(new Date())),
                fxSourceAmount: tx.fxSourceAmount,
                fxSourceCurrency: tx.fxSourceCurrency,
                internalReferenceNumbers: tx.internalReferenceNumbers
            })
        );
    }

    ngOnDestroy(): void {
        this.destroy$.next();
        this.destroy$.complete();
    }
}
