import { TableState } from './table-query-state.helper';

const KEY_PREFIX = 'table-state:';
const TTL_MS = 24 * 60 * 60 * 1000;

interface StoredEntry {
    state: TableState;
    expiresAt: number;
}

export class TableStatePersistence {
    static read(page: string, sessionId: string): TableState | null {
        const key = this.key(page, sessionId);
        const raw = this.getItem(key);
        if (!raw) return null;
        let parsed: StoredEntry | null = null;
        try {
            parsed = JSON.parse(raw) as StoredEntry;
        } catch {
            this.removeItem(key);
            return null;
        }
        if (!parsed || typeof parsed.expiresAt !== 'number' || parsed.expiresAt < Date.now()) {
            this.removeItem(key);
            return null;
        }
        return parsed.state ?? null;
    }

    static write(page: string, sessionId: string, state: TableState): void {
        const key = this.key(page, sessionId);
        const entry: StoredEntry = { state, expiresAt: Date.now() + TTL_MS };
        try {
            localStorage.setItem(key, JSON.stringify(entry));
        } catch {
            return;
        }
        this.gc();
    }

    static clear(page: string, sessionId: string): void {
        this.removeItem(this.key(page, sessionId));
    }

    private static gc(): void {
        const now = Date.now();
        const toDrop: string[] = [];
        try {
            for (let i = 0; i < localStorage.length; i++) {
                const k = localStorage.key(i);
                if (!k || !k.startsWith(KEY_PREFIX)) continue;
                const raw = localStorage.getItem(k);
                if (!raw) continue;
                try {
                    const entry = JSON.parse(raw) as StoredEntry;
                    if (typeof entry.expiresAt !== 'number' || entry.expiresAt < now) {
                        toDrop.push(k);
                    }
                } catch {
                    toDrop.push(k);
                }
            }
        } catch {
            return;
        }
        for (const k of toDrop) this.removeItem(k);
    }

    private static key(page: string, sessionId: string): string {
        return `${KEY_PREFIX}${page}:${sessionId}`;
    }

    private static getItem(key: string): string | null {
        try {
            return localStorage.getItem(key);
        } catch {
            return null;
        }
    }

    private static removeItem(key: string): void {
        try {
            localStorage.removeItem(key);
        } catch {
        }
    }
}
