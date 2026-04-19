import { ImportSource } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/import/v1/import_pb';

export interface TransactionsImportConfig {
    excludedImporters: ImportSource[];
}

export const TRANSACTIONS_IMPORT_PAGE_ID = 'transactions-import';

export const TRANSACTIONS_IMPORT_DEFAULTS: TransactionsImportConfig = Object.freeze({
    excludedImporters: Object.freeze([] as ImportSource[]),
}) as unknown as TransactionsImportConfig;
