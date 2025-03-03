import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { ConfigService } from '../config.service';
import { CookieService } from '../cookie.service';
import { GrpcClientFactoryService } from '../../modules/common/grpc/grpc-client-factory.service';
import { BaseGrpcConnectService } from '../../modules/common/grpc/common/base-grpc-connect.service';
import { AccountsService } from '../../../../gen/gomoneypb/accounts/v1/accounts_connect';
import { Observable } from 'rxjs';
import {
  CreateCurrencyRequest, CreateCurrencyResponse, DeleteCurrencyRequest, DeleteCurrencyResponse,
  ExchangeRequest,
  ExchangeResponse,
  GetCurrenciesRequest, GetCurrenciesResponse, UpdateCurrencyRequest, UpdateCurrencyResponse
} from '../../../../gen/gomoneypb/currency/v1/currency_pb';

@Injectable({
  providedIn: 'root'
})
export class CurrenciesGrpcService extends BaseGrpcConnectService {
  private readonly router = inject(Router);

  constructor(private configurationService: ConfigService,
              private cookieService: CookieService,
              private httpClient: HttpClient,
              grpcClientFactoryService: GrpcClientFactoryService) {
    super(grpcClientFactoryService, configurationService.getConfig().UsersService, AccountsService);
  }

  public exchange(request: ExchangeRequest): Observable<ExchangeResponse> {
    return this.send('Exchange',
      {
        request
      },
      false
    )
  }

  public getCurrencies(request: GetCurrenciesRequest): Observable<GetCurrenciesResponse> {
    return this.send('GetCurrencies',
      {
        request
      },
      false
    )
  }

  public createCurrency(request: CreateCurrencyRequest): Observable<CreateCurrencyResponse> {
    return this.send('CreateCurrency',
      {
        request
      },
      false
    )
  }

  public updateCurrency(request: UpdateCurrencyRequest): Observable<UpdateCurrencyResponse> {
    return this.send('UpdateCurrency',
      {
        request
      },
      false
    )
  }

  public deleteCurrency(request: DeleteCurrencyRequest): Observable<DeleteCurrencyResponse> {
    return this.send('DeleteCurrency',
      {
        request
      },
      false
    )
  }
}
