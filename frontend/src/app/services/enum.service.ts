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
}

export class EnumService {
    static getImportTypes(): AccountTypeEnum[] {
        return [{
            name: "Firefly",
            value: ImportSource.FIREFLY
        }]
    }
    static getAccountTypes(): AccountTypeEnum[] {
        return [
            {
                name: "Regular",
                value: AccountType.REGULAR
            },
            {
                name: "Savings",
                value: AccountType.SAVINGS
            },
            {
                name: "Brokerage",
                value: AccountType.BROKERAGE
            },
            {
                name: "Liability",
                value: AccountType.LIABILITY
            }
        ];
    }

    static getBaseTransactionTypes(): AccountTypeEnum[] {
        return [
            {
                name: "Withdrawal",
                value: TransactionType.WITHDRAWAL
            },
            {
                name: "Deposit",
                value: TransactionType.DEPOSIT
            },
            {
                name: "Transfer",
                value: TransactionType.TRANSFER_BETWEEN_ACCOUNTS
            }
        ]
    }

    static getAllTransactionTypes(): AccountTypeEnum[] {
        return [
            ...EnumService.getBaseTransactionTypes(),
            {
                name: "Reconciliation",
                value: TransactionType.RECONCILIATION
            }
        ];
    }
}
