import { inject } from '@angular/core';
import { Interceptor } from "@connectrpc/connect";
import { CookieService } from '../app/services/cookie.service';
import { CookieInstances } from '../app/objects/cookie-instances';

export function authInterceptor(): Interceptor {
    const cookiesService = inject(CookieService);

    return (next) => async (req) => {
        let jwtValue = cookiesService.get(CookieInstances.Jwt)

        if (!jwtValue) {
            return await next(req);
        }

        req.header.set('Authorization',  `Bearer ${jwtValue}`);

        return await next(req);
    };
}
