import { inject } from '@angular/core';
import { Interceptor } from "@connectrpc/connect";
import { CookieService } from '../app/services/cookie.service';

export function authInterceptor(): Interceptor {
    const cookiesService = inject(CookieService);

    return (next) => async (req) => {
        console.log(cookiesService)
        console.log(`sending message to ${req.url}`);
        return await next(req);
    };
}
