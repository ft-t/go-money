import { Router } from '@angular/router';

export class ReturnUrlHelper {
    /** Full current url including query + fragment, suitable for `returnUrl` query param. */
    static build(router: Router): string {
        return router.url;
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
