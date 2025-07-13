import { Injectable } from '@angular/core';
import { LRUCache } from 'lru-cache';

@Injectable()
export class CacheService {
    private cache = new LRUCache<string, any>({
        max: 500,
        ttl: 5 * 60 * 1000
    });

    public get(key: string): any {
        return this.cache.get(key);
    }

    public set(key: string, value: any) {
        this.cache.set(key, value);
    }

    public clear() {
        this.cache.clear();
    }
}
