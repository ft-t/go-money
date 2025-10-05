import { Injectable } from '@angular/core';
import { LRUCache } from 'lru-cache';

class Cache {
    private cache: LRUCache<string, any>;

    constructor(ttl: number) {
        this.cache = new LRUCache<string, any>({
            max: 500,
            ttl: ttl
        });
    }

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

@Injectable()
export class ShortLivedCache extends Cache {
    constructor() {
        super(1000); // 1 second
    }
}

@Injectable()
export class DefaultCache extends Cache {
    constructor() {
        super(5 * 60 * 1000); // 5 minutes
    }
}
