import { create } from '@bufbuild/protobuf';
import {
    CreateTransactionRequestSchema,
    CreateTransactionResponseSchema,
    UpdateTransactionResponseSchema
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { TransactionSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { TransactionSaveSession, TransactionSaveWriter } from './transaction-save-session';

function createWriter(): jasmine.SpyObj<TransactionSaveWriter> {
    return jasmine.createSpyObj<TransactionSaveWriter>('TransactionSaveWriter', ['createTransaction', 'updateTransaction']);
}

function createRequest(title: string) {
    return create(CreateTransactionRequestSchema, { title });
}

describe('TransactionSaveSession', () => {
    it('stores returned identity and skips unchanged create on retry', async () => {
        const writer = createWriter();

        writer.createTransaction.and.resolveTo(
            create(CreateTransactionResponseSchema, {
                transaction: create(TransactionSchema, { id: 41n })
            })
        );
        const session = new TransactionSaveSession(writer);
        const request = createRequest('first');
        const applied: bigint[] = [];

        await session.save([{ id: 0n, request }], (_, transaction) => applied.push(transaction.id));
        const retryResult = await session.save([{ id: 41n, request }], (_, transaction) => applied.push(transaction.id));

        expect(writer.createTransaction).toHaveBeenCalledTimes(1);
        expect(writer.updateTransaction).not.toHaveBeenCalled();
        expect(applied).toEqual([41n]);
        expect(retryResult).toEqual({ savedCount: 1, discardedCount: 0 });
        expect(session.getCardState(0, request)).toBe('saved');
    });

    it('reports a rule-discarded create as discarded instead of failing', async () => {
        const writer = createWriter();

        writer.createTransaction.and.resolveTo(create(CreateTransactionResponseSchema, { discarded: true }));
        const session = new TransactionSaveSession(writer);
        const request = createRequest('bcd');
        const applied: bigint[] = [];

        const result = await session.save([{ id: 0n, request }], (_, transaction) => applied.push(transaction.id));

        expect(result).toEqual({ savedCount: 0, discardedCount: 1 });
        expect(applied).toEqual([]);
        expect(session.getCardState(0, request)).toBe('discarded');
    });

    it('does not resend a discarded create when the request is unchanged', async () => {
        const writer = createWriter();

        writer.createTransaction.and.resolveTo(create(CreateTransactionResponseSchema, { discarded: true }));
        const session = new TransactionSaveSession(writer);
        const request = createRequest('bcd');

        await session.save([{ id: 0n, request }], () => undefined);
        const retryResult = await session.save([{ id: 0n, request }], () => undefined);

        expect(writer.createTransaction).toHaveBeenCalledTimes(1);
        expect(retryResult).toEqual({ savedCount: 1, discardedCount: 0 });
    });

    it('updates a previously saved transaction after its request changes', async () => {
        const writer = createWriter();

        writer.createTransaction.and.resolveTo(
            create(CreateTransactionResponseSchema, {
                transaction: create(TransactionSchema, { id: 41n })
            })
        );
        writer.updateTransaction.and.resolveTo(
            create(UpdateTransactionResponseSchema, {
                transaction: create(TransactionSchema, { id: 41n, title: 'changed' })
            })
        );
        const session = new TransactionSaveSession(writer);

        await session.save([{ id: 0n, request: createRequest('first') }], () => undefined);
        const changedRequest = createRequest('changed');
        const result = await session.save([{ id: 41n, request: changedRequest }], () => undefined);

        expect(result).toEqual({ savedCount: 1, discardedCount: 0 });
        expect(writer.updateTransaction).toHaveBeenCalledTimes(1);
        expect(writer.updateTransaction.calls.mostRecent().args[0].id).toBe(41n);
        expect(session.getCardState(0, changedRequest)).toBe('saved');
    });

    it('stops at first failure and leaves later transactions pending', async () => {
        const writer = createWriter();

        writer.createTransaction.and.rejectWith(new Error('invalid transaction'));
        const session = new TransactionSaveSession(writer);
        const first = createRequest('first');
        const second = createRequest('second');

        const result = await session.save(
            [
                { id: 0n, request: first },
                { id: 0n, request: second }
            ],
            () => undefined
        );

        expect(result.savedCount).toBe(0);
        expect(result.failedIndex).toBe(0);
        expect(result.error).toEqual(jasmine.any(Error));
        expect(writer.createTransaction).toHaveBeenCalledTimes(1);
        expect(session.getCardState(0, first)).toBe('failed');
        expect(session.getCardState(1, second)).toBe('pending');
    });

    it('rejects create response without assigned identity', async () => {
        const writer = createWriter();

        writer.createTransaction.and.resolveTo(
            create(CreateTransactionResponseSchema, {
                transaction: create(TransactionSchema, {})
            })
        );
        const session = new TransactionSaveSession(writer);

        const result = await session.save([{ id: 0n, request: createRequest('first') }], () => undefined);

        expect(result.failedIndex).toBe(0);
        expect((result.error as Error).message).toBe('Create transaction response is missing an assigned transaction ID');
    });

    it('rejects update response without transaction', async () => {
        const writer = createWriter();

        writer.updateTransaction.and.resolveTo(create(UpdateTransactionResponseSchema, {}));
        const session = new TransactionSaveSession(writer);

        const result = await session.save([{ id: 41n, request: createRequest('first') }], () => undefined);

        expect(result.failedIndex).toBe(0);
        expect((result.error as Error).message).toBe('Update transaction response is missing a transaction');
    });

    it('rejects update response with mismatched transaction identity', async () => {
        const writer = createWriter();

        writer.updateTransaction.and.resolveTo(
            create(UpdateTransactionResponseSchema, {
                transaction: create(TransactionSchema, { id: 99n })
            })
        );
        const session = new TransactionSaveSession(writer);

        const result = await session.save([{ id: 41n, request: createRequest('first') }], () => undefined);

        expect(result.failedIndex).toBe(0);
        expect((result.error as Error).message).toBe('Update transaction response ID does not match requested transaction ID');
    });

    it('ignores concurrent save while one is active', async () => {
        const writer = createWriter();
        let finishCreate!: (value: ReturnType<typeof create<typeof CreateTransactionResponseSchema>>) => void;

        writer.createTransaction.and.returnValue(
            new Promise((resolve) => {
                finishCreate = resolve;
            })
        );
        const session = new TransactionSaveSession(writer);
        const request = createRequest('first');

        const activeSave = session.save([{ id: 0n, request }], () => undefined);
        const concurrentResult = await session.save([{ id: 0n, request }], () => undefined);

        finishCreate(
            create(CreateTransactionResponseSchema, {
                transaction: create(TransactionSchema, { id: 41n })
            })
        );
        await activeSave;

        expect(concurrentResult).toEqual({ savedCount: 0, discardedCount: 0 });
        expect(writer.createTransaction).toHaveBeenCalledTimes(1);
    });

    it('removes saved state at same index as deleted draft', async () => {
        const writer = createWriter();

        writer.createTransaction.and.resolveTo(
            create(CreateTransactionResponseSchema, {
                transaction: create(TransactionSchema, { id: 41n })
            })
        );
        const session = new TransactionSaveSession(writer);
        const first = createRequest('first');
        const second = createRequest('second');

        await session.save([{ id: 0n, request: first }], () => undefined);
        session.remove(0);

        expect(session.getCardState(0, second)).toBe('pending');
    });

    it('resets every saved snapshot and card state', async () => {
        const writer = createWriter();

        writer.createTransaction.and.resolveTo(
            create(CreateTransactionResponseSchema, {
                transaction: create(TransactionSchema, { id: 41n })
            })
        );
        const session = new TransactionSaveSession(writer);
        const request = createRequest('first');

        await session.save([{ id: 0n, request }], () => undefined);

        session.reset();

        expect(session.getCardState(0, request)).toBe('pending');
    });
});
