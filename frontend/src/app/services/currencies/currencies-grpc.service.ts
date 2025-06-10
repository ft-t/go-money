import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { ConfigService } from '../config.service';
import { CookieService } from '../cookie.service';

@Injectable({
  providedIn: 'root'
})
export class CurrenciesGrpcService {
  private readonly router = inject(Router);

  constructor(private configurationService: ConfigService,
              private cookieService: CookieService,
              private httpClient: HttpClient) {
  }
  //
  // public exchange(request: ExchangeRequest): Observable<ExchangeResponse> {
  //   return this.send('Exchange',
  //     {
  //       request
  //     },
  //     false
  //   )
  // }
  //
  // public getCurrencies(request: GetCurrenciesRequest): Observable<GetCurrenciesResponse> {
  //   return this.send('GetCurrencies',
  //     {
  //       request
  //     },
  //     false
  //   )
  // }
  //
  // public createCurrency(request: CreateCurrencyRequest): Observable<CreateCurrencyResponse> {
  //   return this.send('CreateCurrency',
  //     {
  //       request
  //     },
  //     false
  //   )
  // }
  //
  // public updateCurrency(request: UpdateCurrencyRequest): Observable<UpdateCurrencyResponse> {
  //   return this.send('UpdateCurrency',
  //     {
  //       request
  //     },
  //     false
  //   )
  // }
  //
  // public deleteCurrency(request: DeleteCurrencyRequest): Observable<DeleteCurrencyResponse> {
  //   return this.send('DeleteCurrency',
  //     {
  //       request
  //     },
  //     false
  //   )
  // }
}
