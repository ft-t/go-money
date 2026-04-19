import { AfterViewInit, Component, ElementRef, Inject, OnInit, ViewChild } from '@angular/core';
import { Table, TableModule } from 'primeng/table';
import { FormsModule } from '@angular/forms';
import { InputText, InputTextModule } from 'primeng/inputtext';
import { ToastModule } from 'primeng/toast';
import { InputIcon, InputIconModule } from 'primeng/inputicon';
import { IconField, IconFieldModule } from 'primeng/iconfield';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { Transport, createClient } from '@connectrpc/connect';
import { AccountsService, ListAccountsResponse_AccountItem } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { ErrorHelper } from '../../helpers/error.helper';
import { FilterMetadata, MessageService, SortMeta } from 'primeng/api';
import { CommonModule, DatePipe } from '@angular/common';
import { TimestampHelper } from '../../helpers/timestamp.helper';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { Button, ButtonModule } from 'primeng/button';
import { EnumService, AccountTypeEnum } from '../../services/enum.service';
import { MultiSelectModule } from 'primeng/multiselect';
import { SelectModule } from 'primeng/select';
import { OverlayModule } from 'primeng/overlay';
import { Currency, CurrencySchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/currency_pb';
import { create } from '@bufbuild/protobuf';
import { DialogModule } from 'primeng/dialog';
import { InputGroup } from 'primeng/inputgroup';
import { InputGroupAddon } from 'primeng/inputgroupaddon';
import { InputNumber } from 'primeng/inputnumber';
import { ReconciliationModalComponent } from '../transactions/modals/reconciliation-modal/reconciliation-modal.component';
import { TooltipModule } from 'primeng/tooltip';
import { AnalyticsService, GetDebitsAndCreditsSummaryRequestSchema, GetDebitsAndCreditsSummaryResponse_SummaryItem } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/analytics/v1/analytics_pb';
import { TimestampSchema } from '@bufbuild/protobuf/wkt';
import { SelectedDateService } from '../../core/services/selected-date.service';
import { ConfigurationService, GetConfigurationResponse, GetConfigurationResponseSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/configuration/v1/configuration_pb';
import { combineLatest, skip } from 'rxjs';
import { Tag } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/tag_pb';
import { TagsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/tags/v1/tags_pb';
import { FancyTagComponent } from '../../shared/components/fancy-tag/fancy-tag.component';
import { PageConfigService } from '../../services/page-config.service';
import {
    AccountsListConfig,
    ACCOUNTS_LIST_DEFAULTS,
    ACCOUNTS_LIST_PAGE_ID,
    QuickTag,
} from './accounts-list.config';
import { ReturnUrlHelper } from '../../shared/helpers/return-url.helper';
import { TableQueryStateHelper } from '../../shared/helpers/table-query-state.helper';
import { TableStatePersistence } from '../../shared/helpers/table-state-persistence.helper';
import { TabSessionService } from '../../shared/services/tab-session.service';

@Component({
    selector: 'app-account-list',
    templateUrl: 'accounts-list.component.html',
    imports: [
        OverlayModule,
        FormsModule,
        InputTextModule,
        ToastModule,
        TableModule,
        InputIconModule,
        IconFieldModule,
        ButtonModule,
        MultiSelectModule,
        SelectModule,
        CommonModule,
        RouterLink,
        DialogModule,
        ReconciliationModalComponent,
        TooltipModule,
        FancyTagComponent,
    ],
    styles: `
        :host ::ng-deep .accountListTable .p-datatable-header {
            border-width: 0 !important;
        }
    `
})
export class AccountsListComponent implements OnInit, AfterViewInit {
    @ViewChild('dt1', { static: false }) table!: Table;

    statuses: any[] = [];

    loading: boolean = false;
    public reconciliationDialogVisible = false;

    public accountTypesMap: { [id: string]: AccountTypeEnum } = {};

    public accounts: ListAccountsResponse_AccountItem[] = [];
    public tags: Tag[] = [];
    public tagsMap: { [id: number]: Tag } = {};
    public selectedTagIds: number[] = [];
    private accountService;
    private analyticsService;
    private configService;
    private tagsService;
    public accountTypes = EnumService.getAccountTypes();
    public filters: { [s: string]: FilterMetadata } = {};
    public accountCurrencies: Currency[] = [];
    public selectedAccount: ListAccountsResponse_AccountItem | undefined = undefined;
    public multiSortMeta: SortMeta[] = [
        {
            field: 'account.displayOrder',
            order: 0
        }
    ];
    @ViewChild('filter') filter!: ElementRef;
    public analyticsMap: { [accountId: number]: GetDebitsAndCreditsSummaryResponse_SummaryItem } = {};
    public serverConfig: GetConfigurationResponse = create(GetConfigurationResponseSchema, {});
    public pageConfig: AccountsListConfig = { ...ACCOUNTS_LIST_DEFAULTS };
    public editingQuickTags = false;
    public newQuickTagLabel = '';
    public newQuickTagTagIds: number[] = [];
    public editingIndex: number | null = null;
    public editDraft: { label: string; tagIds: number[] } | null = null;
    public initialGlobalFilter: string = '';

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        public router: Router,
        route: ActivatedRoute,
        private selectedDateService: SelectedDateService,
        private pageConfigService: PageConfigService,
        private tabSession: TabSessionService
    ) {
        this.accountService = createClient(AccountsService, this.transport);
        this.analyticsService = createClient(AnalyticsService, this.transport);
        this.configService = createClient(ConfigurationService, this.transport);
        this.tagsService = createClient(TagsService, this.transport);

        if (route.snapshot.data['filters']) {
            for (let ob of route.snapshot.data['filters']) {
                for (let [key, value] of Object.entries(ob)) {
                    this.filters[key] = value as FilterMetadata;
                }
            }
        }

        if (route.snapshot.queryParamMap.get('restore') === '1') {
            const stored = TableStatePersistence.read(this.stateKey, this.tabSession.id);
            if (stored) {
                if (stored.filters) this.filters = { ...this.filters, ...(stored.filters as { [s: string]: FilterMetadata }) };
                if (stored.sort && stored.sort.length > 0) this.multiSortMeta = stored.sort;
                if (stored.global) this.initialGlobalFilter = stored.global;
                const tagIds = stored.extra?.['tagIds'];
                if (Array.isArray(tagIds)) this.selectedTagIds = tagIds as number[];
            }
            TableStatePersistence.clear(this.stateKey, this.tabSession.id);
            this.router.navigate([], { relativeTo: route, queryParams: { restore: null }, queryParamsHandling: 'merge', replaceUrl: true });
        }

        const queryState = TableQueryStateHelper.decode(route.snapshot.queryParams);
        if (queryState.filters) {
            this.filters = { ...this.filters, ...(queryState.filters as { [s: string]: FilterMetadata }) };
        }
        if (queryState.sort && queryState.sort.length > 0) {
            this.multiSortMeta = queryState.sort;
        }
        if (queryState.global) {
            this.initialGlobalFilter = queryState.global;
        }
    }

    private readonly stateKey = 'accounts';

    ngAfterViewInit() {
        if (this.initialGlobalFilter && this.table) {
            if (this.filter?.nativeElement) {
                this.filter.nativeElement.value = this.initialGlobalFilter;
            }
            this.table.filterGlobal(this.initialGlobalFilter, 'contains');
        }
    }

    syncStateToUrl(): void {
        if (!this.table) return;
        const globalVal = (this.table.filters as any)?.['global']?.value;
        TableStatePersistence.write(this.stateKey, this.tabSession.id, {
            filters: this.table.filters as { [f: string]: FilterMetadata | FilterMetadata[] },
            sort: this.table.multiSortMeta ?? [],
            global: typeof globalVal === 'string' ? globalVal : undefined,
            extra: { tagIds: this.selectedTagIds },
        });
    }

    getAccountUrl(account: ListAccountsResponse_AccountItem): string {
        return this.router.createUrlTree(['/', 'accounts', account.account!.id.toString()]).toString();
    }

    public get currentReturnUrl(): string {
        return ReturnUrlHelper.build(this.router);
    }

    async ngOnInit() {
        await this.loadTags();
        await this.loadConfig();
        await this.loadPageConfig();
        await this.loadAccounts();
        await this.loadAnalytics();

        combineLatest([
            this.selectedDateService.fromDate,
            this.selectedDateService.toDate
        ]).pipe(
            skip(1)
        ).subscribe(async () => {
            await this.loadAnalytics();
        });
    }

    async loadTags() {
        try {
            const resp = await this.tagsService.listTags({});
            this.tags = (resp.tags || []).filter(t => !!t?.tag).map(t => t.tag!);
            this.tagsMap = {};
            for (const tag of this.tags) {
                this.tagsMap[tag.id] = tag;
            }
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    getTag(tagID: number): Tag | undefined {
        return this.tagsMap[tagID];
    }

    async onTagFilterChange() {
        this.syncStateToUrl();
        await this.loadAccounts();
        await this.loadAnalytics();
    }

    async loadAccounts() {
        this.loading = true;

        for (let type of this.accountTypes) {
            this.accountTypesMap[type.value] = type;
        }

        let foundCurrencies: { [s: string]: boolean } = {};
        this.accountCurrencies = [];

        try {
            let resp = await this.accountService.listAccounts({ tagIds: this.selectedTagIds });
            this.accounts = resp.accounts || [];

            for (let account of this.accounts) {
                if (account.account && account.account.currency && !foundCurrencies[account.account.currency]) {
                    foundCurrencies[account.account.currency] = true;
                    this.accountCurrencies.push(
                        create(CurrencySchema, {
                            id: account.account.currency
                        })
                    );
                }
            }
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        } finally {
            this.loading = false;
        }
    }

    onGlobalFilter(table: Table, event: Event) {
        table.filterGlobal((event.target as HTMLInputElement).value, 'contains');
    }

    clear(table: Table) {
        table.clear();
        this.filter.nativeElement.value = '';
    }

    showReconcile(account: ListAccountsResponse_AccountItem) {
        this.selectedAccount = account;
        this.reconciliationDialogVisible = true;
    }

    async refreshTable() {
        if (!this.table) {
            return;
        }

        await this.loadAccounts();
        await this.loadAnalytics();
        console.log("Refreshing table data");

        this.table.filter('', '', '');
    }

    async loadConfig() {
        try {
            this.serverConfig = await this.configService.getConfiguration({});
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    async loadPageConfig(): Promise<void> {
        try {
            this.pageConfig = await this.pageConfigService.get<AccountsListConfig>(
                ACCOUNTS_LIST_PAGE_ID,
                ACCOUNTS_LIST_DEFAULTS,
            );
        } catch (e) {
            console.error('Failed to load accounts-list page config:', e);
            this.pageConfig = { ...ACCOUNTS_LIST_DEFAULTS };
        }
    }

    applyQuickTag(tag: QuickTag): void {
        this.selectedTagIds = [...(tag.tagIds ?? [])];
        if (this.filter?.nativeElement) {
            this.filter.nativeElement.value = '';
        }
        if (this.table) {
            this.table.filterGlobal('', 'contains');
        }
        void this.onTagFilterChange();
    }

    async addQuickTag(): Promise<void> {
        const label = this.newQuickTagLabel.trim();
        const tagIds = this.newQuickTagTagIds;
        if (!label || tagIds.length === 0) {
            return;
        }
        this.pageConfig = {
            ...this.pageConfig,
            quickTags: [...this.pageConfig.quickTags, { label, tagIds: [...tagIds] }],
        };
        this.newQuickTagLabel = '';
        this.newQuickTagTagIds = [];
        await this.savePageConfig();
    }

    getQuickTagNames(tag: QuickTag): string {
        if (!tag.tagIds || tag.tagIds.length === 0) {
            return '(no tags)';
        }
        return tag.tagIds
            .map(id => this.getTag(id)?.name ?? `#${id}`)
            .join(', ');
    }

    async removeQuickTag(index: number): Promise<void> {
        if (this.editingIndex !== null) {
            this.editingIndex = null;
            this.editDraft = null;
        }
        this.pageConfig = {
            ...this.pageConfig,
            quickTags: this.pageConfig.quickTags.filter((_, i) => i !== index),
        };
        await this.savePageConfig();
    }

    beginEditQuickTag(index: number): void {
        const row = this.pageConfig.quickTags[index];
        if (!row) {
            return;
        }
        this.editingIndex = index;
        this.editDraft = {
            label: row.label,
            tagIds: [...(row.tagIds ?? [])],
        };
    }

    cancelEditQuickTag(): void {
        this.editingIndex = null;
        this.editDraft = null;
    }

    async saveEditQuickTag(): Promise<void> {
        if (this.editingIndex === null || !this.editDraft) {
            return;
        }
        const label = this.editDraft.label.trim();
        const tagIds = this.editDraft.tagIds;
        if (!label || tagIds.length === 0) {
            return;
        }
        const index = this.editingIndex;
        this.pageConfig = {
            ...this.pageConfig,
            quickTags: this.pageConfig.quickTags.map((row, i) =>
                i === index ? { label, tagIds: [...tagIds] } : row,
            ),
        };
        this.editingIndex = null;
        this.editDraft = null;
        await this.savePageConfig();
    }

    private async savePageConfig(): Promise<void> {
        try {
            await this.pageConfigService.set(ACCOUNTS_LIST_PAGE_ID, this.pageConfig);
        } catch (e) {
            console.error('Failed to save accounts-list page config:', e);
            this.messageService.add({
                severity: 'error',
                summary: 'Save failed',
                detail: ErrorHelper.getMessage(e),
            });
        }
    }

    async loadAnalytics() {
        if (this.accounts.length === 0) {
            return;
        }

        try {
            const fromDate = this.selectedDateService.fromDate.value;
            const toDate = this.selectedDateService.toDate.value;
            const accountIds = this.accounts.map(a => a.account!.id);

            const response = await this.analyticsService.getDebitsAndCreditsSummary(
                create(GetDebitsAndCreditsSummaryRequestSchema, {
                    accountIds: accountIds,
                    startAt: create(TimestampSchema, TimestampHelper.dateToTimestamp(fromDate)),
                    endAt: create(TimestampSchema, TimestampHelper.dateToTimestamp(toDate))
                })
            );

            this.analyticsMap = response.items;
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    getAnalytics(accountId: number): GetDebitsAndCreditsSummaryResponse_SummaryItem | undefined {
        return this.analyticsMap[accountId];
    }

    formatAmount(amount: number): string {
        return parseFloat(amount.toString()).toFixed(2);
    }

    getFilteredAccounts(): ListAccountsResponse_AccountItem[] {
        // Use filtered accounts if available (when user applies filters), otherwise use all accounts
        return this.table?.filteredValue || this.accounts;
    }

    getUniqueCurrencies(): string[] {
        const currencies = new Set<string>();
        const filteredAccounts = this.getFilteredAccounts();
        for (let accountItem of filteredAccounts) {
            if (accountItem.account && accountItem.account.currency) {
                currencies.add(accountItem.account.currency);
            }
        }
        return Array.from(currencies).sort();
    }

    getTotalBalanceByCurrency(currency: string): number {
        let total = 0;
        const filteredAccounts = this.getFilteredAccounts();
        for (let accountItem of filteredAccounts) {
            if (accountItem.account &&
                accountItem.account.currency === currency &&
                accountItem.account.currentBalance) {
                total += parseFloat(accountItem.account.currentBalance);
            }
        }
        return total;
    }

    getTotalDebitsByCurrency(currency: string): number {
        let total = 0;
        const filteredAccounts = this.getFilteredAccounts();
        for (let accountItem of filteredAccounts) {
            if (accountItem.account &&
                accountItem.account.currency === currency) {
                const analytics = this.analyticsMap[accountItem.account.id];
                if (analytics) {
                    total += parseFloat(analytics.totalDebitsAmount);
                }
            }
        }
        return total;
    }

    getTotalCreditsByCurrency(currency: string): number {
        let total = 0;
        const filteredAccounts = this.getFilteredAccounts();
        for (let accountItem of filteredAccounts) {
            if (accountItem.account &&
                accountItem.account.currency === currency) {
                const analytics = this.analyticsMap[accountItem.account.id];
                if (analytics) {
                    total += parseFloat(analytics.totalCreditsAmount);
                }
            }
        }
        return total;
    }

    protected readonly TimestampHelper = TimestampHelper;
}
