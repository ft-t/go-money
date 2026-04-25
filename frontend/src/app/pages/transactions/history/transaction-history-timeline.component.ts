import { Component, Inject, Input, OnChanges, OnInit, SimpleChanges } from '@angular/core';
import { MessageService } from 'primeng/api';
import { Timeline } from 'primeng/timeline';
import { Tag } from 'primeng/tag';
import { RouterLink } from '@angular/router';
import {
    TransactionHistoryActorType,
    TransactionHistoryEvent,
    TransactionHistoryEventType
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/history/v1/history_pb';
import { RulesService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/rules/v1/rules_pb';
import { Rule, ScheduleRule } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/rule_pb';
import { TransactionHistoryClient } from '../../../services/transaction-history.service';
import { ErrorHelper } from '../../../helpers/error.helper';
import { TimestampHelper } from '../../../helpers/timestamp.helper';
import { TRANSPORT_TOKEN } from '../../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';

interface DiffOp {
    op: string;
    path: string;
    value?: unknown;
}

type TagSeverity = 'success' | 'info' | 'warn' | 'danger' | 'secondary' | 'contrast';

@Component({
    selector: 'app-transaction-history-timeline',
    standalone: true,
    imports: [Timeline, Tag, RouterLink],
    templateUrl: './transaction-history-timeline.component.html'
})
export class TransactionHistoryTimelineComponent implements OnChanges, OnInit {
    @Input() transactionId!: bigint;

    public events: TransactionHistoryEvent[] = [];
    public loading = false;

    private rulesById = new Map<number, Rule>();
    private scheduleRulesById = new Map<number, ScheduleRule>();
    private rulesService;

    public readonly ActorType = TransactionHistoryActorType;

    constructor(
        @Inject(TRANSPORT_TOKEN) private readonly transport: Transport,
        private readonly historyClient: TransactionHistoryClient,
        private readonly messageService: MessageService
    ) {
        this.rulesService = createClient(RulesService, this.transport);
    }

    async ngOnInit() {
        await Promise.all([this.fetchRules(), this.fetchScheduleRules()]);
    }

    ngOnChanges(changes: SimpleChanges) {
        if (changes['transactionId'] && this.transactionId !== undefined && this.transactionId !== null) {
            void this.fetch();
        }
    }

    private async fetch() {
        this.loading = true;
        try {
            const events = await this.historyClient.listHistory(this.transactionId);
            this.events = [...events].reverse();
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            this.events = [];
        } finally {
            this.loading = false;
        }
    }

    private async fetchRules() {
        try {
            const resp = await this.rulesService.listRules({});
            for (const r of resp.rules || []) {
                this.rulesById.set(r.id, r);
            }
        } catch {
            // silent — actor link just falls back to "Rule #N"
        }
    }

    private async fetchScheduleRules() {
        try {
            const resp = await this.rulesService.listScheduleRules({});
            for (const r of resp.rules || []) {
                this.scheduleRulesById.set(r.id, r);
            }
        } catch {
            // silent — actor link falls back to "Scheduled #N"
        }
    }

    ruleName(id: number | undefined): string {
        if (id === undefined) return 'Rule';
        const r = this.rulesById.get(id);
        return r?.title ? r.title : `Rule #${id}`;
    }

    ruleLink(id: number | undefined): (string | number)[] | null {
        if (id === undefined) return null;
        return ['/', 'rules', 'edit', id.toString()];
    }

    scheduleName(id: number | undefined): string {
        if (id === undefined) return 'Scheduler';
        const r = this.scheduleRulesById.get(id);
        return r?.title ? r.title : `Scheduled #${id}`;
    }

    scheduleLink(id: number | undefined): (string | number)[] | null {
        if (id === undefined) return null;
        return ['/', 'schedules', 'edit', id.toString()];
    }

    actorLabel(event: TransactionHistoryEvent): string {
        switch (event.actorType) {
            case TransactionHistoryActorType.USER:
                return event.actorUserId !== undefined ? `User #${event.actorUserId}` : 'User';
            case TransactionHistoryActorType.RULE:
                return this.ruleName(event.actorRuleId);
            case TransactionHistoryActorType.SCHEDULER:
                return this.scheduleName(event.actorRuleId);
            case TransactionHistoryActorType.IMPORTER:
                return `Importer: ${event.actorExtra || 'unknown'}`;
            case TransactionHistoryActorType.BULK:
                const op = event.actorExtra || 'op';
                const user = event.actorUserId !== undefined ? `user #${event.actorUserId}` : 'unknown user';
                return `Bulk (${op}) by ${user}`;
            case TransactionHistoryActorType.UNSPECIFIED:
            default:
                return 'Unknown';
        }
    }

    eventTypeLabel(type: TransactionHistoryEventType): string {
        switch (type) {
            case TransactionHistoryEventType.CREATED:
                return 'Created';
            case TransactionHistoryEventType.UPDATED:
                return 'Updated';
            case TransactionHistoryEventType.DELETED:
                return 'Deleted';
            case TransactionHistoryEventType.RULE_APPLIED:
                return 'Rule applied';
            case TransactionHistoryEventType.UNSPECIFIED:
            default:
                return 'Unknown';
        }
    }

    eventTypeIcon(type: TransactionHistoryEventType): string {
        switch (type) {
            case TransactionHistoryEventType.CREATED:
                return 'pi pi-plus';
            case TransactionHistoryEventType.UPDATED:
                return 'pi pi-pencil';
            case TransactionHistoryEventType.DELETED:
                return 'pi pi-trash';
            case TransactionHistoryEventType.RULE_APPLIED:
                return 'pi pi-cog';
            default:
                return 'pi pi-circle';
        }
    }

    eventTypeColor(type: TransactionHistoryEventType): string {
        switch (type) {
            case TransactionHistoryEventType.CREATED:
                return '#22c55e';
            case TransactionHistoryEventType.UPDATED:
                return '#f59e0b';
            case TransactionHistoryEventType.DELETED:
                return '#ef4444';
            case TransactionHistoryEventType.RULE_APPLIED:
                return '#6366f1';
            default:
                return '#64748b';
        }
    }

    diffOps(event: TransactionHistoryEvent): DiffOp[] {
        const diff = event.diff;
        if (!diff) return [];
        const raw = (diff as Record<string, unknown>)['ops'];
        if (!Array.isArray(raw)) return [];
        return raw.filter((op): op is DiffOp => !!op && typeof op === 'object' && 'op' in op && 'path' in op);
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

    eventTimestamp(event: TransactionHistoryEvent): string {
        if (!event.occurredAt) return '';
        const d = TimestampHelper.timestampToDate(event.occurredAt);
        return d.toLocaleString();
    }

    trackEvent(_: number, event: TransactionHistoryEvent): string {
        return event.id.toString();
    }
}
