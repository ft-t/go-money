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
}
