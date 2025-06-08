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
import { GrpcClientFactoryService } from './app/modules/common/grpc/grpc-client-factory.service';

export const appConfig: ApplicationConfig = {
  providers: [
    provideRouter(appRoutes, withInMemoryScrolling({
      anchorScrolling: 'enabled',
      scrollPositionRestoration: 'enabled'
    }), withEnabledBlockingInitialNavigation()),
    provideHttpClient(withFetch()),
    provideAnimationsAsync(),
    providePrimeNG({theme: {preset: Aura, options: {darkModeSelector: '.app-dark'}}}),
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
    BroadcastService,
    {
      provide: GrpcClientFactoryService,
      useClass: GrpcClientFactoryService
    },
  ]
};
