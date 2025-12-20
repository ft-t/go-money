import { Transaction } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { toJson, fromJson } from '@bufbuild/protobuf';
import { TransactionSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';

export interface Snippet {
    id: string;
    name: string;
    transactions: unknown[];
    order: number;
    createdAt: string;
    updatedAt: string;
}

export function serializeTransactions(transactions: Transaction[]): unknown[] {
    return transactions.map(tx => toJson(TransactionSchema, tx));
}

export function deserializeTransactions(data: unknown[]): Transaction[] {
    return data.map(item => fromJson(TransactionSchema, item as any));
}
