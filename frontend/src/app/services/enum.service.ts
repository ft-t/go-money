import { Injectable } from '@angular/core';
import { AccountType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';

@Injectable({
    providedIn: 'root'
})

export class AccountTypeEnum {
    name: string = '';
    value: number = 0;
}

export class EnumService {
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
}
