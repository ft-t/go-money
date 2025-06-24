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
import { authInterceptor } from './app/core/interceptors/auth';
import { MessageService } from 'primeng/api';
import { EnumService } from './app/services/enum.service';
import { SelectedDateService } from './app/core/services/selected-date.service';
import { DatePipe } from '@angular/common';
import { CookieService } from './app/services/cookie.service';
import { CookieInstances } from './app/objects/cookie-instances';
import { provideHighlightOptions } from 'ngx-highlightjs';

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
        provideHighlightOptions({
            fullLibraryLoader: () => import('highlight.js'),
        }),

        provideHttpClient(withFetch()),
        provideAnimationsAsync(),
        MessageService,
        EnumService,
        SelectedDateService,
        DatePipe,
        providePrimeNG({ theme: { preset: Aura, options: { darkModeSelector: '.app-dark' } } }),
        {
            provide: TRANSPORT_TOKEN,
            deps: [CookieService],
            useFactory: (cookiesService: CookieService) => {
                let host = cookiesService.get(CookieInstances.CustomApiHost);
                if (!host) host = '/';

                return createConnectTransport({
                    baseUrl: host,
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
