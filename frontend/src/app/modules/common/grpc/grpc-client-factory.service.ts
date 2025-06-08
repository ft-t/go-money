import { Injectable } from '@angular/core';
import { IGrpcClientFactoryService } from './i-grpc-client-factory.service';
import { createConnectTransport } from '@connectrpc/connect-web';
import { createPromiseClient, PromiseClient } from '@connectrpc/connect';
import { BroadcastService } from '../../../services/broadcast.service';
import { CookieService } from '../../../services/cookie.service';
import { ConfigService } from '../../../services/config.service';
import { CookieInstances } from '../../../objects/cookie-instances';

@Injectable()
export class GrpcClientFactoryService implements IGrpcClientFactoryService {

  constructor(public broadcastService: BroadcastService,
              private cookieService: CookieService,
              private readonly config: ConfigService,) {
  }

  createClient(host: string, service: any): PromiseClient<any> {
    const userAgentInterceptor: any = (next: any) => async (req: any) => {

      try {
        const header2 = req.header.get('user-agent');

        if (header2 && header2.startsWith('connect-es')) {
          req.header.set('user-agent', navigator.userAgent);
        }
      } catch (ex) {
        console.log(ex);
        return await next(req);
      }

      return await next(req);
    };

    let useBinaryFormat = true;
    try {
      const useBinaryFormatCookie = this.cookieService.get(CookieInstances.UseBinaryFormat, null);

      if (useBinaryFormatCookie) {
        useBinaryFormat = JSON.parse(useBinaryFormatCookie) as boolean;
      }
    } catch (e) {}

    const transport = createConnectTransport({
      baseUrl: host,
      useBinaryFormat: false,
      jsonOptions: {
        ignoreUnknownFields: true
      },
      interceptors: [userAgentInterceptor],
    });

    return createPromiseClient(service, transport);
  }

  getJwt(): string {
    return this.cookieService.get(CookieInstances.Jwt);
  }

  getXSeed(): string {
    return '';
  }

  hasDebug(): boolean {
    return !!this.cookieService.get(CookieInstances.DebugGrpcInConsole);
  }
}

