import { provideHttpClient, withFetch } from '@angular/common/http';
import { ApplicationConfig } from '@angular/core';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { provideRouter, withEnabledBlockingInitialNavigation, withInMemoryScrolling } from '@angular/router';
import Aura from '@primeng/themes/aura';
import { providePrimeNG } from 'primeng/config';
import { appRoutes } from './app.routes';
import { ConfigService } from './app/services/config.service';
import { InitialConfiguration } from './app/objects/configuration/сonfiguration';
import { BroadcastService } from './app/services/broadcast.service';
import { createConnectTransport } from '@connectrpc/connect-web';
import { TRANSPORT_TOKEN } from './app/consts/transport';
import { authInterceptor } from './interceptors/auth';
import { MessageService } from 'primeng/api';
import { EnumService } from './app/services/enum.service';

export const appConfig: ApplicationConfig = {
    providers: [
        provideRouter(
            appRoutes,
            withInMemoryScrolling({
                anchorScrolling: 'enabled',
                scrollPositionRestoration: 'enabled'
            }),
            withEnabledBlockingInitialNavigation()
        ),
        provideHttpClient(withFetch()),
        provideAnimationsAsync(),
        MessageService,
        EnumService,
        providePrimeNG({ theme: { preset: Aura, options: { darkModeSelector: '.app-dark' } } }),
        {
            provide: TRANSPORT_TOKEN,
            useFactory: () => {
                return createConnectTransport({
                    baseUrl: 'http://localhost:52055',
                    interceptors: [authInterceptor()]
                });
            }
        },
        {
            provide: ConfigService,
            useFactory: () => {
                let configData: any = {};
                // tslint:disable-next-line:no-non-null-assertion
                try {
                    configData = JSON.parse(document!.head.getAttribute('content')!);
                } catch (ex) {
                    //
                }

                const initial = configData ? new InitialConfiguration(configData.Environment) : new InitialConfiguration('prod');

                return new ConfigService(initial);
            },
            deps: []
        },
        BroadcastService
    ]
};
