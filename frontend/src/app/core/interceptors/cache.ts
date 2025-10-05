import { Interceptor } from '@connectrpc/connect';
import { inject } from '@angular/core';
import { CookieService } from '../../services/cookie.service';
import { Router } from '@angular/router';
import { MessageService } from 'primeng/api';
import { DefaultCache, ShortLivedCache } from '../services/cache.service';
import { Rule } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/rule_pb';
import { Mutex } from 'async-mutex';

type CacheInstance = DefaultCache | ShortLivedCache;

export function cacheInterceptor(): Interceptor {
    const defaultCache = inject(DefaultCache);
    const shortLivedCache = inject(ShortLivedCache);

    const cacheableMethods: { [key: string]: { cache: CacheInstance } } = {
        "ListTags": { cache: defaultCache },
        "ListCategories": { cache: defaultCache },
        "GetConfiguration": { cache: defaultCache },
        "GetApplicableAccounts": { cache: shortLivedCache },
        "GetCurrencies": { cache: shortLivedCache },
    };

    const mutexes = new Map<string, Mutex>();

    return (next) => async (req) => {
        const targetPath = req.method.name ?? 'unk';

        if (!cacheableMethods[targetPath] || req.stream) {
            return await next(req);
        }

        const config = cacheableMethods[targetPath];
        const key = req.url + JSON.stringify(req.message);

        const cachedData = config.cache.get(key);
        if (cachedData) {
            return cachedData;
        }

        if (!mutexes.has(key)) {
            mutexes.set(key, new Mutex());
        }
        const mutex = mutexes.get(key)!;

        const release = await mutex.acquire();

        try {
            const cachedDataAfterLock = config.cache.get(key);
            if (cachedDataAfterLock) {
                return cachedDataAfterLock;
            }

            const resp = await next(req);
            config.cache.set(key, resp);

            return resp;
        } finally {
            release();
        }
    };
}
