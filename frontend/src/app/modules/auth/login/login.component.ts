import { ChangeDetectionStrategy, Component, OnDestroy, OnInit } from '@angular/core';
import { BaseAutoUnsubscribeClass } from '../../../objects/auto-unsubscribe/base-auto-unsubscribe-class';
import { UsersGrpcService } from '../../../services/auth/users-grpc.service';
import { createConnectTransport } from '@connectrpc/connect-web';
import { createClient } from '@connectrpc/connect';

import { CreateRequest, CreateResponse, LoginRequest, LoginResponse } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/users/v1/users_pb';
import { Router } from '@angular/router';
import { CookieService } from '../../../services/cookie.service';
import { ConfigurationService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/configuration/v1/configuration_pb';

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
export class LoginComponent implements OnInit {
    email: string = '';
    password: string = '';

    authType = AuthTypeEnum.Login;

    protected readonly AuthTypeEnum = AuthTypeEnum;

    constructor(
        private usersService: UsersGrpcService,
        private cookieService: CookieService,
        private router: Router
    ) {}

    async ngOnInit() {

    }

    changeAuthType(type: AuthTypeEnum) {
        this.authType = type;
    }

    async login() {
        let transport = createConnectTransport({
            baseUrl: 'http://localhost:8080',
        });

        const client = createClient(ConfigurationService, transport);
        let val = await client.getConfiguration({});

        // const request = new LoginRequest({
        //     login: this.email,
        //     password: this.password
        // });

        // this.usersService.loginMethod(request)
        //   .pipe(
        //     tap((response: LoginResponse) => {
        //       console.log(response);
        //
        //       this.cookieService.set(CookieInstances.Jwt, '12345', new Date(2099, 1, 1), '/');
        //
        //       this.router.navigate(['/']);
        //     }),
        //     this.takeUntilDestroy
        //   ).subscribe();
    }

    register() {
        // const request = new CreateRequest({
        //     login: this.email,
        //     password: this.password
        // });

        // this.usersService.register(request)
        //   .pipe(
        //     tap((response: CreateResponse) => {
        //       console.log(response);
        //
        //       this.cookieService.set(CookieInstances.Jwt, '12345', new Date(2099, 1, 1), '/');
        //
        //       this.router.navigate(['/']);
        //     }),
        //     this.takeUntilDestroy
        //   ).subscribe();
    }
}
