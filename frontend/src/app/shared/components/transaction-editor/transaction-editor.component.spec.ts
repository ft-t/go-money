import { FormControl, FormGroup } from '@angular/forms';
import { create } from '@bufbuild/protobuf';
import { TransactionSchema, TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { TransactionEditorComponent } from './transaction-editor.component';

describe('TransactionEditorComponent transaction drafts', () => {
    it('exports every editable live form value without server identity', () => {
        const editor = Object.create(TransactionEditorComponent.prototype) as TransactionEditorComponent;
        editor.form = new FormGroup({
            id: new FormControl(99n),
            sourceAmount: new FormControl('12.34'),
            sourceCurrency: new FormControl('PLN'),
            sourceAccountId: new FormControl(11),
            destinationAmount: new FormControl('3.21'),
            destinationCurrency: new FormControl('EUR'),
            destinationAccountId: new FormControl(22),
            notes: new FormControl('note'),
            title: new FormControl('Clone me'),
            categoryId: new FormControl(7),
            transactionDate: new FormControl(new Date('2026-07-12T10:30:00.000Z')),
            type: new FormControl(TransactionType.EXPENSE),
            tagIds: new FormControl([2, 5]),
            skipRules: new FormControl(true),
            fxSourceAmount: new FormControl('13.45'),
            fxSourceCurrency: new FormControl('USD'),
            internalReferenceNumbers: new FormControl(['reference'])
        });

        const draft = editor.buildTransactionDraft();

        expect(draft.transaction.id).toBe(0n);
        expect(draft.transaction.sourceAmount).toBe('12.34');
        expect(draft.transaction.sourceCurrency).toBe('PLN');
        expect(draft.transaction.sourceAccountId).toBe(11);
        expect(draft.transaction.destinationAmount).toBe('3.21');
        expect(draft.transaction.destinationCurrency).toBe('EUR');
        expect(draft.transaction.destinationAccountId).toBe(22);
        expect(draft.transaction.notes).toBe('note');
        expect(draft.transaction.title).toBe('Clone me');
        expect(draft.transaction.categoryId).toBe(7);
        expect(draft.transaction.transactionDate).toBeDefined();
        expect(draft.transaction.type).toBe(TransactionType.EXPENSE);
        expect(draft.transaction.tagIds).toEqual([2, 5]);
        expect(draft.transaction.fxSourceAmount).toBe('13.45');
        expect(draft.transaction.fxSourceCurrency).toBe('USD');
        expect(draft.transaction.internalReferenceNumbers).toEqual(['reference']);
        expect(draft.skipRules).toBeTrue();
    });

    it('seeds skip rules from cloned draft state', () => {
        const editor = Object.create(TransactionEditorComponent.prototype) as TransactionEditorComponent;
        editor.initialSkipRules = true;
        editor.accounts = {};

        const form = editor.buildForm(create(TransactionSchema, {}));

        expect(form.get('skipRules')!.value).toBeTrue();
    });
});
