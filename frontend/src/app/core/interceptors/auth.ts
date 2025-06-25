import { inject } from '@angular/core';
import { Code, ConnectError, Interceptor } from '@connectrpc/connect';
import { CookieService } from '../../services/cookie.service';
import { CookieInstances } from '../../objects/cookie-instances';
import { Router } from '@angular/router';
import { MessageService } from 'primeng/api';
import { ErrorHelper } from '../../helpers/error.helper';

export function authInterceptor(): Interceptor {
    const cookiesService = inject(CookieService);
    const router = inject(Router);
    const messageService = inject(MessageService);

    return (next) => async (req) => {
        let jwtValue = cookiesService.get(CookieInstances.Jwt);

        if (!jwtValue) {
            return await next(req);
        }

        if (req.method?.localName != "login") {
            req.header.set('Authorization', `Bearer ${jwtValue}`);
        }

        try {
            let resp = await next(req);
            return resp;
        } catch (e) {
            console.log(e);
            messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });

            if (e instanceof ConnectError) {
                if (e.code == Code.Unauthenticated) {
                    cookiesService.delete(CookieInstances.Jwt, '/');
                    await router.navigate(['/', 'login']);
                }
            }
            throw e;
        }
    };
}
