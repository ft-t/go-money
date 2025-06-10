import { ChangeDetectionStrategy, Component, Inject, OnDestroy, OnInit } from '@angular/core';
import { BaseAutoUnsubscribeClass } from '../../../objects/auto-unsubscribe/base-auto-unsubscribe-class';
import { UsersGrpcService } from '../../../services/auth/users-grpc.service';
import { createConnectTransport } from '@connectrpc/connect-web';
import { createClient, Transport } from '@connectrpc/connect';

import { Router } from '@angular/router';
import { CookieService } from '../../../services/cookie.service';
import { ConfigurationService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/configuration/v1/configuration_pb';
import { FormsModule } from '@angular/forms';
import { Password } from 'primeng/password';
import { Button } from 'primeng/button';
import { AppFloatingConfigurator } from '../../../layout/component/app.floatingconfigurator';
import { InputText } from 'primeng/inputtext';
import { TRANSPORT_TOKEN } from '../../../consts/transport';
import { NgIf } from '@angular/common';
import { UsersService, CreateRequest, CreateRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/users/v1/users_pb';
import { CreateAccountRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';

export enum AuthTypeEnum {
    Login = 0,
    Register = 1
}

@Component({
    selector: 'app-login',
    templateUrl: 'login.component.html',
    imports: [FormsModule, Password, Button, AppFloatingConfigurator, InputText, NgIf]
})
export class LoginComponent implements OnInit {
    email: string = '';
    password: string = '';

    isRegisterFlow: boolean = false;

    authType = AuthTypeEnum.Login;

    private configService;
    private usersService;

    protected readonly AuthTypeEnum = AuthTypeEnum;

    constructor(
        private cookieService: CookieService,
        private router: Router,
        @Inject(TRANSPORT_TOKEN) private transport: Transport
    ) {
        this.configService = createClient(ConfigurationService, this.transport);
        this.usersService = createClient(UsersService, this.transport);
    }

    async ngOnInit() {
        let val = await this.configService.getConfiguration({});

        this.isRegisterFlow = val.shouldCreateAdmin;
    }

    changeAuthType(type: AuthTypeEnum) {
        this.authType = type;
    }

    async login() {
        let val = await this.configService.getConfiguration({});
    }

    async register() {
        let xx = CreateRequestSchema;

        await this.usersService.create({
            login: this.email,
            password: this.password
        });
    }
}
