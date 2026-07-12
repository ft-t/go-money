import { FormControl, FormGroup, Validators } from '@angular/forms';
import { Subject } from 'rxjs';
import { create } from '@bufbuild/protobuf';
import {
    CreateTransactionRequestSchema,
    CreateTransactionResponseSchema
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { TransactionSchema, TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { ReturnUrlHelper } from '../../shared/helpers/return-url.helper';
import { TransactionEditorComponent, TransactionDraft } from '../../shared/components/transaction-editor/transaction-editor.component';
import { TransactionSaveSession, TransactionSaveWriter } from './transaction-save-session';
import { TransactionUpsertComponent } from './transactions-upsert.component';

interface EditorDouble {
    getForm: jasmine.Spy<() => FormGroup>;
    buildTransactionRequest: jasmine.Spy<() => ReturnType<typeof create<typeof CreateTransactionRequestSchema>>>;
    buildTransactionDraft: jasmine.Spy<() => TransactionDraft>;
}

function createWriter(): jasmine.SpyObj<TransactionSaveWriter> {
    return jasmine.createSpyObj<TransactionSaveWriter>('TransactionSaveWriter', ['createTransaction', 'updateTransaction']);
}

function createEditor(title: string, valid = true): EditorDouble {
    const form = new FormGroup({
        title: new FormControl(valid ? title : '', Validators.required)
    });
    const request = create(CreateTransactionRequestSchema, { title });
    const draft = {
        transaction: create(TransactionSchema, { title, type: TransactionType.EXPENSE }),
        skipRules: true
    };

    return {
        getForm: jasmine.createSpy('getForm').and.returnValue(form),
        buildTransactionRequest: jasmine.createSpy('buildTransactionRequest').and.returnValue(request),
        buildTransactionDraft: jasmine.createSpy('buildTransactionDraft').and.returnValue(draft)
    };
}

function createComponents(editors: EditorDouble[]) {
    return {
        get: (index: number) => editors[index] as unknown as TransactionEditorComponent,
        [Symbol.iterator]: () => editors[Symbol.iterator]()
    };
}

function createComponent(writer = createWriter()) {
    const component = Object.create(TransactionUpsertComponent.prototype) as TransactionUpsertComponent;
    const messageService = jasmine.createSpyObj('MessageService', ['add']);

    Object.assign(component, {
        targetTransaction: [create(TransactionSchema, {})],
        initialSkipRules: [false],
        components: createComponents([]),
        saveSession: new TransactionSaveSession(writer),
        messageService,
        router: {},
        route: {}
    });

    return { component, messageService, writer };
}

describe('TransactionUpsertComponent', () => {
    it('adds a fully blank transaction', () => {
        const { component } = createComponent();

        component.addTransaction();

        expect(component.targetTransaction[1]).toEqual(create(TransactionSchema, {}));
        expect(component.initialSkipRules[1]).toBeFalse();
    });

    it('blocks draft actions while save requests are active', () => {
        const { component } = createComponent();

        component.saveSession.isSaving = true;

        component.addTransaction();

        expect(component.targetTransaction.length).toBe(1);
        expect(component.initialSkipRules).toEqual([false]);
    });

    it('clones current live editor draft', () => {
        const { component } = createComponent();
        const editor = createEditor('Clone me');

        component.components = createComponents([editor]) as never;

        component.cloneTransaction(0);

        expect(component.targetTransaction[1].id).toBe(0n);
        expect(component.targetTransaction[1].title).toBe('Clone me');
        expect(component.targetTransaction[1].type).toBe(TransactionType.EXPENSE);
        expect(component.initialSkipRules[1]).toBeTrue();
    });

    it('validates every editor before sending requests', async () => {
        const { component, messageService, writer } = createComponent();

        component.components = createComponents([createEditor('invalid', false)]) as never;

        await component.saveAll();

        expect(writer.createTransaction).not.toHaveBeenCalled();
        expect(messageService.add).toHaveBeenCalledWith({
            severity: 'error',
            detail: 'Please fix validation errors in all transactions'
        });
    });

    it('applies successful identity and stops at failed transaction', async () => {
        const { component, messageService, writer } = createComponent();

        component.targetTransaction.push(create(TransactionSchema, { title: 'second' }));
        component.initialSkipRules.push(false);
        component.components = createComponents([createEditor('first'), createEditor('second')]) as never;
        writer.createTransaction.and.resolveTo(
            create(CreateTransactionResponseSchema, {
                transaction: create(TransactionSchema, { id: 41n, title: 'first' })
            })
        );
        writer.createTransaction.and.rejectWith(new Error('invalid amount'));
        writer.createTransaction.and.returnValues(
            Promise.resolve(
                create(CreateTransactionResponseSchema, {
                    transaction: create(TransactionSchema, { id: 41n, title: 'first' })
                })
            ),
            Promise.reject(new Error('invalid amount'))
        );
        spyOn(ReturnUrlHelper, 'navigateAfterSave').and.resolveTo();

        await component.saveAll();

        expect(component.targetTransaction[0].id).toBe(41n);
        expect(component.targetTransaction[1].title).toBe('second');
        expect(writer.createTransaction).toHaveBeenCalledTimes(2);
        expect(ReturnUrlHelper.navigateAfterSave).not.toHaveBeenCalled();
        expect(messageService.add).toHaveBeenCalledWith({
            severity: 'error',
            detail: '1 transaction saved. Transaction 2 failed: invalid amount'
        });
    });

    it('deletes aligned draft and save state', () => {
        const { component } = createComponent();

        component.targetTransaction.push(create(TransactionSchema, { title: 'second' }));
        component.initialSkipRules.push(true);
        const remove = spyOn(component.saveSession, 'remove');

        component.deleteDraft(1);

        expect(component.targetTransaction.length).toBe(1);
        expect(component.targetTransaction[0].title).toBe('');
        expect(component.initialSkipRules).toEqual([false]);
        expect(remove).toHaveBeenCalledOnceWith(1);
    });

    it('resets aligned client state when snippets replace every draft', () => {
        const { component } = createComponent();

        component.initialSkipRules = [true, true];
        const reset = spyOn(component.saveSession, 'reset');

        component.applySnippetTransactions([
            create(TransactionSchema, { title: 'one' }),
            create(TransactionSchema, { title: 'two' })
        ]);

        expect(component.targetTransaction.map((transaction) => transaction.title)).toEqual(['one', 'two']);
        expect(component.initialSkipRules).toEqual([false, false]);
        expect(reset).toHaveBeenCalledTimes(1);
    });

    it('keeps skip rule state aligned when commission adds a draft', async () => {
        const { component } = createComponent();
        const originalForm = new FormGroup({
            transactionDate: new FormControl(new Date('2026-07-12T10:30:00.000Z'))
        });
        const editor = {
            getForm: jasmine.createSpy('getForm').and.returnValue(originalForm),
            adjustSourceAmount: jasmine.createSpy('adjustSourceAmount')
        };

        component.components = createComponents([editor as never]) as never;
        component.expenseSplitForm = new FormGroup({
            sourceAccountId: new FormControl(11),
            sourceCurrency: new FormControl('PLN'),
            destinationAccountId: new FormControl(22),
            destinationCurrency: new FormControl('PLN'),
            amount: new FormControl('5'),
            title: new FormControl('Commission', Validators.required)
        });
        Object.assign(component, {
            currentSplitIndex: 0,
            destroy$: new Subject<void>(),
            showExpenseSplit: true
        });

        await component.saveExpenseSplit();

        expect(component.targetTransaction.length).toBe(2);
        expect(component.initialSkipRules).toEqual([false, false]);
    });

    it('disables editor forms while save requests are active', async () => {
        const { component, writer } = createComponent();
        const editor = createEditor('first');
        let finishCreate!: (response: ReturnType<typeof create<typeof CreateTransactionResponseSchema>>) => void;

        component.components = createComponents([editor]) as never;
        writer.createTransaction.and.returnValue(
            new Promise((resolve) => {
                finishCreate = resolve;
            })
        );
        spyOn(ReturnUrlHelper, 'navigateAfterSave').and.resolveTo();

        const save = component.saveAll();

        expect(editor.getForm().disabled).toBeTrue();
        finishCreate(
            create(CreateTransactionResponseSchema, {
                transaction: create(TransactionSchema, { id: 41n, title: 'first' })
            })
        );
        await save;

        expect(editor.getForm().enabled).toBeTrue();
    });

    it('navigates after every transaction saves', async () => {
        const { component, writer } = createComponent();

        component.components = createComponents([createEditor('first')]) as never;
        writer.createTransaction.and.resolveTo(
            create(CreateTransactionResponseSchema, {
                transaction: create(TransactionSchema, { id: 41n, title: 'first' })
            })
        );
        spyOn(ReturnUrlHelper, 'navigateAfterSave').and.resolveTo();

        await component.saveAll();

        expect(ReturnUrlHelper.navigateAfterSave).toHaveBeenCalled();
    });
});
