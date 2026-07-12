import { clone, create, equals } from '@bufbuild/protobuf';
import {
    CreateTransactionRequest,
    CreateTransactionRequestSchema,
    CreateTransactionResponse,
    UpdateTransactionRequest,
    UpdateTransactionRequestSchema,
    UpdateTransactionResponse
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { Transaction } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';

export type TransactionCardState = 'pending' | 'saved' | 'unsaved' | 'failed';

export interface TransactionSaveWriter {
    createTransaction(request: CreateTransactionRequest): Promise<CreateTransactionResponse>;
    updateTransaction(request: UpdateTransactionRequest): Promise<UpdateTransactionResponse>;
}

export interface TransactionSaveItem {
    id: bigint;
    request: CreateTransactionRequest;
}

export interface TransactionSaveResult {
    savedCount: number;
    failedIndex?: number;
    error?: unknown;
}

export class TransactionSaveSession {
    public isSaving = false;
    private savedRequests: Array<CreateTransactionRequest | undefined> = [];
    private states: TransactionCardState[] = [];

    constructor(private readonly writer: TransactionSaveWriter) {}

    async save(items: TransactionSaveItem[], apply: (index: number, transaction: Transaction) => void): Promise<TransactionSaveResult> {
        if (this.isSaving) {
            return { savedCount: 0 };
        }

        this.isSaving = true;
        this.states = this.states.map((state) => (state === 'failed' ? 'pending' : state));
        let savedCount = 0;
        let currentIndex = 0;

        try {
            for (currentIndex = 0; currentIndex < items.length; currentIndex++) {
                const item = items[currentIndex];
                const savedRequest = this.savedRequests[currentIndex];
                if (savedRequest && equals(CreateTransactionRequestSchema, savedRequest, item.request)) {
                    continue;
                }

                let transaction: Transaction | undefined;
                if (item.id === 0n) {
                    const response = await this.writer.createTransaction(item.request);
                    transaction = response.transaction;
                    if (!transaction || transaction.id === 0n) {
                        throw new Error('Create transaction response is missing an assigned transaction ID');
                    }
                } else {
                    const response = await this.writer.updateTransaction(
                        create(UpdateTransactionRequestSchema, {
                            id: item.id,
                            transaction: item.request
                        })
                    );
                    transaction = response.transaction;
                    if (!transaction) {
                        throw new Error('Update transaction response is missing a transaction');
                    }
                }

                apply(currentIndex, transaction);
                this.savedRequests[currentIndex] = clone(CreateTransactionRequestSchema, item.request);
                this.states[currentIndex] = 'saved';
                savedCount++;
            }

            return { savedCount };
        } catch (error) {
            this.states[currentIndex] = 'failed';
            return { savedCount, failedIndex: currentIndex, error };
        } finally {
            this.isSaving = false;
        }
    }

    getCardState(index: number, request: CreateTransactionRequest): TransactionCardState {
        const state = this.states[index] ?? 'pending';
        const savedRequest = this.savedRequests[index];
        if (state === 'saved' && savedRequest && !equals(CreateTransactionRequestSchema, savedRequest, request)) {
            return 'unsaved';
        }

        return state;
    }

    remove(index: number): void {
        this.savedRequests.splice(index, 1);
        this.states.splice(index, 1);
    }
}
