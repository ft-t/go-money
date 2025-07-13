import { Interceptor } from '@connectrpc/connect';
import { inject } from '@angular/core';
import { CookieService } from '../../services/cookie.service';
import { Router } from '@angular/router';
import { MessageService } from 'primeng/api';
import { CacheService } from '../services/cache.service';
import { Rule } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/rule_pb';

export function cacheInterceptor(): Interceptor {
    const cacheService = inject(CacheService);
    const cacheableMethods: { [key: string]: any } = {
        "ListTags": {},
        "ListCategories": {},
        "GetConfiguration": {},
    };
    return (next) => async (req) => {
        const targetPath = req.method.name ?? 'unk';

        if (!cacheableMethods[targetPath] || req.stream) {
            return await next(req);
        }

        const key = req.url + JSON.stringify(req.message);

        const cachedData = cacheService.get(key);
        if (cachedData) {
            return cachedData;
        }


        let resp = await next(req);
        cacheService.set(key, resp);

        return resp;
    };
}
