import {
  ChangeDetectionStrategy,
  Component,
  OnDestroy,
  OnInit,
} from '@angular/core';
import { BaseAutoUnsubscribeClass } from '../../../../objects/auto-unsubscribe/base-auto-unsubscribe-class';

@Component({
  selector: 'app-account-upsert',
  standalone: false,
  templateUrl: 'account-upsert.component.html',
  styleUrls: ['account-upsert.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AccountUpsertComponent extends BaseAutoUnsubscribeClass implements OnInit, OnDestroy {
  constructor() {
    super();
  }

  override ngOnInit() {
    super.ngOnInit();
  }

  override ngOnDestroy() {
    super.ngOnDestroy();
  }

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
