import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { ConfigService } from '../config.service';
import { CookieService } from '../cookie.service';

@Injectable({
  providedIn: 'root'
})
export class AccountsGrpcService {
  private readonly router = inject(Router);

  constructor(private configurationService: ConfigService,
              private cookieService: CookieService,
              private httpClient: HttpClient) {
  }
  //
  // public createAcount(request: CreateAccountRequest): Observable<CreateAccountResponse> {
  //   // return this.send('CreateAccount',
  //   //   {
  //   //     request
  //   //   },
  //   //   false
  //   // )
  // }
  //
  // public updateAccount(request: UpdateAccountRequest): Observable<UpdateAccountResponse> {
  //   // return this.send('UpdateAccount',
  //   //   {
  //   //     request
  //   //   },
  //   //   false
  //   // )
  // }
  //
  // public deleteAccount(request: DeleteAccountRequest): Observable<DeleteAccountResponse> {
  //   // return this.send('DeleteAccount',
  //   //   {
  //   //     request
  //   //   },
  //   //   false
  //   // )
  // }
  //
  // public listAccounts(): Observable<ListAccountsResponse> {
  //   // return this.send('ListAccounts',
  //   //   {},
  //   //   false
  //   // )
  // }

  // public reorderAccounts(request: ReorderAccountsRequest): Observable<ReorderAccountsResponse> {
  //   return this.send('ReorderAccounts',
  //     {
  //       request
  //     },
  //     false
  //   )
  // }
}
