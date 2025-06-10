import {
  ChangeDetectionStrategy,
  Component,
  ElementRef,
  OnDestroy,
  OnInit,
  ViewChild
} from '@angular/core';
import { BaseAutoUnsubscribeClass } from '../../../objects/auto-unsubscribe/base-auto-unsubscribe-class';
import { BehaviorSubject, finalize } from 'rxjs';
import { CurrenciesGrpcService } from '../../../services/currencies/currencies-grpc.service';
import { tap } from 'rxjs/operators';

@Component({
  selector: 'app-currency-list',
  standalone: false,
  templateUrl: 'currency-list.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class CurrencyListComponent extends BaseAutoUnsubscribeClass implements OnInit, OnDestroy {
  public currencies$ = new BehaviorSubject<any[] | any>([]);

  public override isLoading$ = new BehaviorSubject<boolean>(false);

  @ViewChild('filter') filter!: ElementRef;

  constructor(private currenciesService: CurrenciesGrpcService) {
    super();
  }

  override ngOnInit() {
    super.ngOnInit();
  }

  override ngOnDestroy() {
    super.ngOnDestroy();
  }

  // getCurrencies() {
  //
  //   this.isLoading$.next(true);
  //
  //   const request = {
  //   }
  //
  //   this.currenciesService.getCurrencies(request)
  //     .pipe(
  //       tap((response: GetCurrenciesResponse | any) => {
  //         console.log(response);
  //         this.currencies$.next(response.data);
  //       }),
  //       finalize(() => this.isLoading$.next(false)),
  //       this.takeUntilDestroy
  //     ).subscribe();
  // }
  //
  // createCurrency() {
  //   this.isLoading$.next(true);
  //
  //   const request = {
  //
  //   }
  //
  //   this.currenciesService.createCurrency(request)
  //     .pipe(
  //       tap(),
  //       finalize(() => this.isLoading$.next(false)),
  //       this.takeUntilDestroy
  //     ).subscribe();
  // }
  //
  // updateCurrency() {
  //   this.isLoading$.next(true);
  //
  //   const request = {
  //
  //   }
  //
  //   this.currenciesService.updateCurrency(request)
  //     .pipe(
  //       tap(),
  //       finalize(() => this.isLoading$.next(false)),
  //       this.takeUntilDestroy
  //     ).subscribe();
  // }
  //
  // deleteCurrency(currency: any) {
  //   this.isLoading$.next(true);
  //
  //   const request = {
  //
  //   }
  //
  //   this.currenciesService.deleteCurrency(request)
  //     .pipe(
  //       tap(),
  //       finalize(() => this.isLoading$.next(false)),
  //       this.takeUntilDestroy
  //     ).subscribe();
  // }
}
