import {
  Component,
  OnDestroy,
  OnInit,
} from '@angular/core';
import { Button } from 'primeng/button';
import { InputText } from 'primeng/inputtext';
import { Fluid } from 'primeng/fluid';

@Component({
    selector: 'app-account-upsert',
    templateUrl: 'account-upsert.component.html',
    styleUrls: ['account-upsert.component.scss'],
    imports: [Button, InputText, Fluid]
})
export class AccountUpsertComponent implements OnInit, OnDestroy {
    constructor() {}

    ngOnInit() {}

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
