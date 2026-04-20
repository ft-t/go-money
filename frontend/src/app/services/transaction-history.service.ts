import { Inject, Injectable } from '@angular/core';
import { createClient, Transport } from '@connectrpc/connect';
import { TRANSPORT_TOKEN } from '../consts/transport';
import {
    TransactionHistoryService,
    ListTransactionHistoryRequestSchema,
    TransactionHistoryEvent
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/history/v1/history_pb';
import { create } from '@bufbuild/protobuf';

@Injectable({ providedIn: 'root' })
export class TransactionHistoryClient {
    private readonly client;

    constructor(@Inject(TRANSPORT_TOKEN) transport: Transport) {
        this.client = createClient(TransactionHistoryService, transport);
    }

    async listHistory(transactionId: bigint): Promise<TransactionHistoryEvent[]> {
        const resp = await this.client.listHistory(
            create(ListTransactionHistoryRequestSchema, { transactionId })
        );
        return resp.events ?? [];
    }
}
