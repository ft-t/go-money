import { FilterMetadata, SortMeta } from 'primeng/api';
import { Params } from '@angular/router';

export interface TableState {
    filters?: { [field: string]: FilterMetadata | FilterMetadata[] };
    sort?: SortMeta[];
    first?: number;
    rows?: number;
    global?: string;
}

export class TableQueryStateHelper {
    /**
     * Decode a route's query param map into table state.
     * Unknown / malformed params are silently dropped; we never throw on a bad URL.
     */
    static decode(params: Params): TableState {
        const state: TableState = {};

        const filtersRaw = params['filters'];
        if (typeof filtersRaw === 'string' && filtersRaw.length > 0) {
            try {
                const parsed = JSON.parse(filtersRaw);
                if (parsed !== null && typeof parsed === 'object' && !Array.isArray(parsed)) {
                    state.filters = parsed;
                }
            } catch {
                // swallow — bad query param, fall back to empty
            }
        }

        const sortRaw = params['sort'];
        if (typeof sortRaw === 'string' && sortRaw.length > 0) {
            try {
                const parsed = JSON.parse(sortRaw);
                if (Array.isArray(parsed) && parsed.every(e => e !== null && typeof e === 'object')) {
                    state.sort = parsed;
                }
            } catch {
                // swallow
            }
        }

        if (params['first'] != null) {
            const n = Number(params['first']);
            if (!Number.isNaN(n)) state.first = n;
        }

        if (params['rows'] != null) {
            const n = Number(params['rows']);
            if (!Number.isNaN(n)) state.rows = n;
        }

        if (typeof params['global'] === 'string') {
            state.global = params['global'];
        }

        return state;
    }

    /**
     * Encode table state into a query param map. Empty / default values emit `null`
     * so `router.navigate(..., { queryParamsHandling: 'merge' })` clears them.
     */
    static encode(state: TableState): Params {
        const out: Params = {};

        const stripped = state.filters ? this.stripEmpty(state.filters) : {};
        out['filters'] = Object.keys(stripped).length > 0 ? JSON.stringify(stripped) : null;

        if (state.sort && state.sort.length > 0) {
            out['sort'] = JSON.stringify(state.sort);
        } else {
            out['sort'] = null;
        }

        out['first'] = state.first != null && state.first !== 0 ? state.first : null;
        out['rows'] = state.rows != null ? state.rows : null;
        out['global'] = state.global && state.global.length > 0 ? state.global : null;

        return out;
    }

    /** Drop filter entries whose value is null/undefined/'' or empty array. */
    private static stripEmpty(
        filters: { [field: string]: FilterMetadata | FilterMetadata[] }
    ): { [field: string]: FilterMetadata | FilterMetadata[] } {
        const isBlank = (m: FilterMetadata): boolean => {
            const v = m.value;
            if (v == null) return true;
            if (typeof v === 'string' && v.length === 0) return true;
            if (Array.isArray(v) && v.length === 0) return true;
            return false;
        };

        const out: { [f: string]: FilterMetadata | FilterMetadata[] } = {};
        for (const [k, v] of Object.entries(filters)) {
            if (v == null) continue;
            if (Array.isArray(v)) {
                const kept = v.filter(m => m != null && !isBlank(m));
                if (kept.length > 0) out[k] = kept;
                continue;
            }
            if (!isBlank(v)) out[k] = v;
        }
        return out;
    }
}
