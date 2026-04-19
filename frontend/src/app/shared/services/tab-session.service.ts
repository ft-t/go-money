import { Injectable } from '@angular/core';

const STORAGE_KEY = 'tabSessionId';

@Injectable({ providedIn: 'root' })
export class TabSessionService {
    private cached: string | null = null;

    get id(): string {
        if (this.cached) return this.cached;
        let existing: string | null = null;
        try {
            existing = sessionStorage.getItem(STORAGE_KEY);
        } catch {
        }
        if (existing) {
            this.cached = existing;
            return existing;
        }
        const generated = this.generate();
        try {
            sessionStorage.setItem(STORAGE_KEY, generated);
        } catch {
        }
        this.cached = generated;
        return generated;
    }

    private generate(): string {
        const cryptoObj = (globalThis as { crypto?: Crypto }).crypto;
        if (cryptoObj && typeof cryptoObj.randomUUID === 'function') {
            return cryptoObj.randomUUID();
        }
        return Math.random().toString(36).slice(2) + Date.now().toString(36);
    }
}
