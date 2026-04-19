import { Router } from '@angular/router';

export class ReturnUrlHelper {
    /** Bare path (no query, no fragment) — consumers restore filter state from localStorage. */
    static build(router: Router): string {
        return router.url.split('?')[0].split('#')[0];
    }

    /**
     * Validate that a returnUrl is same-origin (starts with `/`) and not protocol-relative.
     * Returns the url unchanged if safe, otherwise null.
     */
    static safe(url: string | null | undefined): string | null {
        if (!url) return null;
        if (!url.startsWith('/')) return null;
        if (url.startsWith('//')) return null;
        return url;
    }

    /** Append `?restore=1` marker so the destination list restores persisted table state once. */
    static withRestoreFlag(url: string): string {
        const sep = url.includes('?') ? '&' : '?';
        return url + sep + 'restore=1';
    }
}
