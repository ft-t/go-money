import { ChangeDetectionStrategy, Component, OnDestroy, OnInit } from '@angular/core';
import { BaseAutoUnsubscribeClass } from '../../../objects/auto-unsubscribe/base-auto-unsubscribe-class';
import { UsersGrpcService } from '../../../services/auth/users-grpc.service';
import { tap } from 'rxjs/operators';
import {
  CreateRequest,
  CreateResponse,
  LoginRequest,
  LoginResponse
} from '../../../../../gen/gomoneypb/users/v1/users_pb';
import { Router } from '@angular/router';
import { CookieService } from '../../../services/cookie.service';
import { CookieInstances } from '../../../objects/cookie-instances';

export enum AuthTypeEnum {
  Login = 0,
  Register = 1
}

@Component({
  selector: 'app-login',
  standalone: false,
  templateUrl: 'login.component.html',
  styleUrls: ['login.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class LoginComponent extends BaseAutoUnsubscribeClass implements OnInit, OnDestroy {
  email: string = '';
  password: string = '';
  checked: boolean = false;

  authType = AuthTypeEnum.Login

  protected readonly AuthTypeEnum = AuthTypeEnum;

  constructor(private usersService: UsersGrpcService,
              private cookieService: CookieService,
              private router: Router) {
    super();
  }

  override ngOnInit() {
    super.ngOnInit();
  }

  override ngOnDestroy() {
    super.ngOnDestroy();
  }

  changeAuthType(type: AuthTypeEnum) {
    this.authType = type;
  }

  login() {
    const request = new LoginRequest(
      {
        login: this.email,
        password: this.password
      }
    );

    this.usersService.loginMethod(request)
      .pipe(
        tap((response: LoginResponse) => {
          console.log(response);

          this.cookieService.set(CookieInstances.Jwt, '12345', new Date(2099, 1, 1), '/');

          this.router.navigate(['/']);
        }),
        this.takeUntilDestroy
      ).subscribe();
  }

  register() {
    const request = new CreateRequest(
      {
        login: this.email,
        password: this.password
      }
    );

    this.usersService.register(request)
      .pipe(
        tap((response: CreateResponse) => {
          console.log(response);

          this.cookieService.set(CookieInstances.Jwt, '12345', new Date(2099, 1, 1), '/');

          this.router.navigate(['/']);
        }),
        this.takeUntilDestroy
      ).subscribe();
  }
}
