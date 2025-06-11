import { Component, Inject, OnInit } from '@angular/core';
import { Button } from 'primeng/button';
import { InputText } from 'primeng/inputtext';
import { Fluid } from 'primeng/fluid';
import { DropdownModule } from 'primeng/dropdown';
import { TRANSPORT_TOKEN } from '../../../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { CurrencyService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/currency/v1/currency_pb';
import { ErrorHelper } from '../../../../helpers/error.helper';
import { MessageService } from 'primeng/api';
import { Account, AccountSchema, AccountType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { FormsModule } from '@angular/forms';
import { Currency } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/currency_pb';
import { create } from '@bufbuild/protobuf';
import { EnumService } from '../../../../services/enum.service';
import { NgIf } from '@angular/common';
import { Textarea } from 'primeng/textarea';
import { AccountsService, CreateAccountRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { Router } from '@angular/router';

@Component({
    selector: 'app-account-upsert',
    templateUrl: 'account-upsert.component.html',
    styleUrls: ['account-upsert.component.scss'],
    imports: [Button, InputText, Fluid, DropdownModule, FormsModule, NgIf, Textarea]
})
export class AccountUpsertComponent implements OnInit {
    private currencyService;
    private accountsService;

    public currencies: Currency[] = [];

    public account: Account = create(AccountSchema, {});
    public accountTypes = EnumService.getAccountTypes();

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        private router: Router,
    ) {
        this.currencyService = createClient(CurrencyService, this.transport);
        this.accountsService = createClient(AccountsService, this.transport);
    }

    async ngOnInit() {
        this.account = create(AccountSchema, {});

        try {
            let response = await this.currencyService.getCurrencies({});

            this.currencies = response.currencies || [];
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    async create() {
        try {
            let response = await this.accountsService.createAccount(
                create(CreateAccountRequestSchema, {
                    type: this.account.type,
                    name: this.account.name,
                    currency: this.account.currency,
                    accountNumber: this.account.accountNumber,
                    iban: this.account.iban,
                    note: this.account.note,
                    liabilityPercent: this.account.liabilityPercent,
                    extra: {
                        created_by: 'web'
                    } // todo
                })
            );

            this.messageService.add({ severity: 'info', detail: "New account created" });
            await this.router.navigate(['/', 'accounts', response.account!.id.toString()]);
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    protected readonly AccountType = AccountType;
}
