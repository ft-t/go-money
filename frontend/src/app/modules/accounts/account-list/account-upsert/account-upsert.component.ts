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

@Component({
    selector: 'app-account-upsert',
    templateUrl: 'account-upsert.component.html',
    styleUrls: ['account-upsert.component.scss'],
    imports: [Button, InputText, Fluid, DropdownModule, FormsModule]
})
export class AccountUpsertComponent implements OnInit {
    private currencyService;
    public currencies: Currency[] = [];

    public account: Account = create(AccountSchema, {});
    public accountTypes = EnumService.getAccountTypes();

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        private enumService: EnumService,
    ) {
        this.currencyService = createClient(CurrencyService, this.transport);
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

    ngOnDestroy() {}

    deleteAccount() {
        // const request = new DeleteAccountRequest(
        //   {
        //     id: 12345
        //   }
        // );
        // this.accountsService.deleteAccount(request)
        //   .pipe(
        //     tap((response: DeleteAccountResponse) => {
        //       console.log(response);
        //     }),
        //     this.takeUntilDestroy
        //   ).subscribe();
    }

    updateAccount() {
        // const request = new UpdateAccountRequest(
        //   {
        //     id: 12345,
        //     name: '',
        //     extra: {},
        //     type: AccountType.UNSPECIFIED,
        //     note: '',
        //     liabilityPercent: '',
        //     iban: '',
        //     accountNumber: ''
        //   }
        // );
        // this.accountsService.updateAccount(request)
        //   .pipe(
        //     tap((response: UpdateAccountResponse) => {
        //       console.log(response);
        //       // this.customers1$.next(response.accounts);
        //     }),
        //     this.takeUntilDestroy
        //   ).subscribe();
    }
}
