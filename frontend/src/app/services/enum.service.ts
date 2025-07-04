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
        return [{
            name: 'Firefly',
            value: ImportSource.FIREFLY,
            icon: ''
        }];
    }

    static getAccountTypes(): AccountTypeEnum[] {
        return [
            {
                name: 'Regular',
                value: AccountType.REGULAR,
                icon: ''
            },
            {
                name: 'Savings',
                value: AccountType.SAVINGS,
                icon: ''
            },
            {
                name: 'Brokerage',
                value: AccountType.BROKERAGE,
                icon: ''
            },
            {
                name: 'Liability',
                value: AccountType.LIABILITY,
                icon: ''
            }
        ];
    }

    static getBaseTransactionTypes(): AccountTypeEnum[] {
        return [
            {
                name: 'Deposit',
                value: TransactionType.DEPOSIT,
                icon: 'pi pi-arrow-down text-green-500'
            },
            {
                name: 'Transfer',
                value: TransactionType.TRANSFER_BETWEEN_ACCOUNTS,
                icon: 'pi pi-arrows-h text-blue-500'
            },
            {
                name: 'Withdrawal',
                value: TransactionType.WITHDRAWAL,
                icon: 'pi pi-arrow-up text-red-500'
            },
        ];
    }

    static getAllTransactionTypes(): AccountTypeEnum[] {
        return [
            ...EnumService.getBaseTransactionTypes(),
            {
                name: 'Reconciliation',
                value: TransactionType.RECONCILIATION,
                icon: '',
            }
        ];
    }
}
