import { Account, AccountType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { GetApplicableAccountsResponse_ApplicableRecord } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { EnumService } from '../services/enum.service';

export class AccountHelper {
    static getApplicableAccounts(
        accounts: { [s: number]: GetApplicableAccountsResponse_ApplicableRecord },
        type: TransactionType,
        isSource: boolean
    ): Account[] {
        const applicable = accounts[type];

        if (!applicable) {
            return [];
        }

        const accountList = isSource ? applicable.sourceAccounts || [] : applicable.destinationAccounts || [];

        return AccountHelper.sortByDisplayOrder(accountList);
    }

    static sortByDisplayOrder(accounts: Account[]): Account[] {
        return accounts.sort((a, b) => {
            const orderA = a.displayOrder ?? 999999;
            const orderB = b.displayOrder ?? 999999;
            return orderA - orderB;
        });
    }

    static getAccountById(allAccounts: { [s: number]: Account }, id: number | undefined): Account | null {
        if (!id) return null;
        return allAccounts[id] || null;
    }

    static getAccountTypeName(type: number): string {
        const accountTypes = EnumService.getAccountTypes();
        const accountType = accountTypes.find(t => t.value === type);
        return accountType?.name || 'Unknown';
    }

    static getDefaultAccountByType(
        accounts: { [s: number]: GetApplicableAccountsResponse_ApplicableRecord },
        type: TransactionType,
        isSource: boolean,
        accountType: AccountType
    ): Account | null {
        const availableAccounts = AccountHelper.getApplicableAccounts(accounts, type, isSource);

        const defaultAccount = availableAccounts.find(acc =>
            acc.type === accountType && (acc.flags & BigInt(1)) === BigInt(1)
        );

        return defaultAccount || null;
    }

    static getExpenseAccountByNameOrDefault(
        accounts: { [s: number]: GetApplicableAccountsResponse_ApplicableRecord },
        accountName: string
    ): Account | null {
        const expenseAccounts = AccountHelper.getApplicableAccounts(accounts, TransactionType.EXPENSE, false);

        const namedAccount = expenseAccounts.find(acc =>
            acc.name.toLowerCase() === accountName.toLowerCase()
        );

        if (namedAccount) {
            return namedAccount;
        }

        return AccountHelper.getDefaultAccountByType(accounts, TransactionType.EXPENSE, false, AccountType.EXPENSE);
    }

    static parseFloat(value: string): number {
        return parseFloat(value);
    }
}
