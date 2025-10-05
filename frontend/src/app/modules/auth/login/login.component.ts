import { Component, Inject, OnInit } from '@angular/core';
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
import { UsersService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/users/v1/users_pb';
import { CookieInstances } from '../../../objects/cookie-instances';
import { ToastModule } from 'primeng/toast';
import { MessageService } from 'primeng/api';
import { ErrorHelper } from '../../../helpers/error.helper';
import { DefaultCache, ShortLivedCache } from '../../../core/services/cache.service';

@Component({
    selector: 'app-login',
    templateUrl: 'login.component.html',
    imports: [FormsModule, Password, Button, AppFloatingConfigurator, InputText, NgIf, ToastModule]
})
export class LoginComponent implements OnInit {
    email: string = '';
    password: string = '';

    isRegisterFlow: boolean = false;

    private configService;
    private usersService;

    constructor(
        private cookieService: CookieService,
        private router: Router,
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        private defaultCache: DefaultCache,
        private shortLivedCache: ShortLivedCache
    ) {
        this.configService = createClient(ConfigurationService, this.transport);
        this.usersService = createClient(UsersService, this.transport);
    }

    async ngOnInit() {
        let val = await this.configService.getConfiguration({});

        this.isRegisterFlow = val.shouldCreateAdmin;
    }

    async login() {
        try {
            let resp = await this.usersService.login({
                login: this.email,
                password: this.password
            });

            this.cookieService.set(CookieInstances.Jwt, resp.token, {
                path: '/'
            });

            await this.router.navigate(['/', 'accounts']);
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    async register() {
        await this.usersService.create({
            login: this.email,
            password: this.password
        });

        this.isRegisterFlow = false;

        this.messageService.add({ severity: 'info', detail: 'User created successfully. You can now login.' });
        this.defaultCache.clear();
        this.shortLivedCache.clear();
    }
}
