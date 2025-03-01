import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { ConfigService } from '../config.service';
import { CookieService } from '../cookie.service';
import { GrpcClientFactoryService } from '../../modules/common/grpc/grpc-client-factory.service';
import { BaseGrpcConnectService } from '../../modules/common/grpc/common/base-grpc-connect.service';
import { Observable } from 'rxjs';
import { UsersService } from '../../../../gen/gomoneypb/users/v1/users_connect';
import {
  CreateRequest,
  CreateResponse,
  LoginRequest,
  LoginResponse
} from '../../../../gen/gomoneypb/users/v1/users_pb';
import { CookieInstances } from '../../objects/cookie-instances';

@Injectable({
  providedIn: 'root'
})
export class UsersGrpcService extends BaseGrpcConnectService {
  private readonly router = inject(Router);

  constructor(private configurationService: ConfigService,
              private cookieService: CookieService,
              private httpClient: HttpClient,
              grpcClientFactoryService: GrpcClientFactoryService) {
    super(grpcClientFactoryService, configurationService.getConfig().UsersService, UsersService);
  }

  public loginMethod(request: LoginRequest): Observable<LoginResponse> {
    return this.send('Login',
      {
        request
      },
      false
    )
  }

  public register(request: CreateRequest): Observable<CreateResponse> {
    return this.send('Create',
      {
        request
      },
      false
    )
  }

  public logout() {
    this.cookieService.delete(CookieInstances.Jwt);
    this.router.navigate(['/', 'login']);
  }
}
