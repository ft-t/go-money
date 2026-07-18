import { create } from '@bufbuild/protobuf';
import { TransactionSchema, TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { TransactionItem, TransactionsImportComponent } from './transactions-import.component';

describe('TransactionsImportComponent review header synchronisation', () => {
    it('applies edited editor values to the reviewed item', () => {
        const component = Object.create(TransactionsImportComponent.prototype) as TransactionsImportComponent;
        const item: TransactionItem = {
            transaction: create(TransactionSchema, {
                id: 5n,
                title: 'Parsed title',
                type: TransactionType.EXPENSE,
                sourceAmount: '10.00',
                sourceCurrency: 'USD'
            }),
            selected: true,
            ignored: false,
            discarded: false,
            appliedRules: [],
            hasError: false
        };

        component.syncTransactionFromEditor(
            item,
            create(TransactionSchema, {
                id: 0n,
                title: 'Edited title',
                type: TransactionType.INCOME,
                sourceAmount: '99.99',
                sourceCurrency: 'PLN'
            })
        );

        expect(item.transaction.title).toBe('Edited title');
        expect(item.transaction.type).toBe(TransactionType.INCOME);
        expect(item.transaction.sourceAmount).toBe('-99.99');
        expect(item.transaction.sourceCurrency).toBe('PLN');
    });

    it('keeps the parsed amount sign convention', () => {
        const component = Object.create(TransactionsImportComponent.prototype) as TransactionsImportComponent;
        const item: TransactionItem = {
            transaction: create(TransactionSchema, { sourceAmount: '-492.72', destinationAmount: '492.72' }),
            selected: true,
            ignored: false,
            discarded: false,
            appliedRules: [],
            hasError: false
        };

        component.syncTransactionFromEditor(item, create(TransactionSchema, { sourceAmount: '492.72', destinationAmount: '492.72' }));

        expect(item.transaction.sourceAmount).toBe('-492.72');
        expect(item.transaction.destinationAmount).toBe('492.72');
    });

    it('keeps the reviewed item identity and server id stable', () => {
        const component = Object.create(TransactionsImportComponent.prototype) as TransactionsImportComponent;
        const original = create(TransactionSchema, { id: 5n, title: 'Parsed title' });
        const item: TransactionItem = { transaction: original, selected: true, ignored: false, discarded: false, appliedRules: [], hasError: false };

        component.syncTransactionFromEditor(item, create(TransactionSchema, { id: 0n, title: 'Edited title' }));

        expect(item.transaction).toBe(original);
        expect(item.transaction.id).toBe(5n);
    });
});
