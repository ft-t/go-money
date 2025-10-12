import { Injectable } from '@angular/core';
import { AccountType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { ImportSource } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/import/v1/import_pb';

@Injectable({
    providedIn: 'root'
})
export class AccountTypeEnum {
    name: string = '';
    value: number = 0;
    icon: string = '';
}

export class EnumService {
    static getImportTypes(): AccountTypeEnum[] {
        return [
            {
                name: 'Firefly',
                value: ImportSource.FIREFLY,
                icon: ''
            },
            {
                name: 'Privat24',
                value: ImportSource.PRIVATE_24,
                icon: ''
            },
            {
                name: 'Revolut',
                value: ImportSource.REVOLUT,
                icon: ''
            },
            {
                name: 'Monobank',
                value: ImportSource.MONOBANK,
                icon: ''
            },
            {
                name: 'BNP Paribas Polska',
                value: ImportSource.BNP_PARIBAS_POLSKA,
                icon: ''
            }
        ];
    }

    static getAccountTypes(): AccountTypeEnum[] {
        return [
            {
                name: 'Asset',
                value: AccountType.ASSET,
                icon: ''
            },
            {
                name: 'Liability',
                value: AccountType.LIABILITY,
                icon: ''
            },
            {
                name: 'Income',
                value: AccountType.INCOME,
                icon: ''
            },
            {
                name: 'Expense',
                value: AccountType.EXPENSE,
                icon: ''
            }
        ];
    }

    static getBaseTransactionTypes(): AccountTypeEnum[] {
        return [
            {
                name: 'Deposit',
                value: TransactionType.INCOME,
                icon: 'pi pi-arrow-down text-green-500'
            },
            {
                name: 'Transfer',
                value: TransactionType.TRANSFER_BETWEEN_ACCOUNTS,
                icon: 'pi pi-arrows-h text-blue-500'
            },
            {
                name: 'Expense',
                value: TransactionType.EXPENSE,
                icon: 'pi pi-arrow-up text-red-500'
            }
        ];
    }

    static getAllTransactionTypes(): AccountTypeEnum[] {
        return [
            ...EnumService.getBaseTransactionTypes(),
            {
                name: 'Reconciliation',
                value: TransactionType.ADJUSTMENT,
                icon: ''
            }
        ];
    }
}
