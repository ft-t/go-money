import { Component, Input } from '@angular/core';
import { Tag } from 'primeng/tag';

export interface DiffOp {
    op: string;
    path: string;
    value?: unknown;
}

type TagSeverity = 'success' | 'info' | 'warn' | 'danger' | 'secondary' | 'contrast';

@Component({
    selector: 'app-diff-ops',
    standalone: true,
    imports: [Tag],
    template: `
        @if (ops.length > 0) {
            <div class="flex flex-col gap-1">
                @for (op of ops; track $index) {
                    <div class="flex items-start gap-2 text-sm">
                        <p-tag [value]="op.op" [severity]="opSeverity(op.op)" />
                        <span class="font-mono text-surface-700 dark:text-surface-200">{{ opPathDisplay(op.path) }}</span>
                        @if (op.value !== undefined && op.op !== 'remove') {
                            <span class="text-surface-500">=</span>
                            <span class="font-mono break-all">{{ opValueDisplay(op.value) }}</span>
                        }
                    </div>
                }
            </div>
        }
    `
})
export class DiffOpsComponent {
    @Input() ops: DiffOp[] = [];

    static parse(diff: unknown): DiffOp[] {
        if (!diff) return [];
        const raw = (diff as Record<string, unknown>)['ops'];
        if (!Array.isArray(raw)) return [];
        return raw.filter((op): op is DiffOp => !!op && typeof op === 'object' && 'op' in op && 'path' in op);
    }

    static parseJson(diffJson: string): DiffOp[] {
        if (!diffJson) return [];
        try {
            return DiffOpsComponent.parse(JSON.parse(diffJson));
        } catch {
            return [];
        }
    }

    opPathDisplay(path: string): string {
        if (!path) return '';
        return path.startsWith('/') ? path.slice(1) : path;
    }

    opValueDisplay(value: unknown): string {
        if (value === undefined) return '';
        if (value === null) return 'null';
        if (typeof value === 'string') return value;
        if (typeof value === 'number' || typeof value === 'boolean') return String(value);
        try {
            return JSON.stringify(value);
        } catch {
            return String(value);
        }
    }

    opSeverity(op: string): TagSeverity {
        switch (op) {
            case 'add':
                return 'success';
            case 'replace':
                return 'warn';
            case 'remove':
                return 'danger';
            default:
                return 'secondary';
        }
    }
}
