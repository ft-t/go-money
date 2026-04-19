import { AfterViewInit, Component, ElementRef, Inject, OnInit, ViewChild } from '@angular/core';
import { Table, TableModule } from 'primeng/table';
import { FormsModule } from '@angular/forms';
import { InputText } from 'primeng/inputtext';
import { ToastModule } from 'primeng/toast';
import { InputIcon } from 'primeng/inputicon';
import { IconField } from 'primeng/iconfield';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { Transport, createClient } from '@connectrpc/connect';
import { AccountsService, ListAccountsResponse_AccountItem } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { ErrorHelper } from '../../helpers/error.helper';
import { FilterMetadata, MessageService } from 'primeng/api';
import { DatePipe } from '@angular/common';
import { TimestampHelper } from '../../helpers/timestamp.helper';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { Button } from 'primeng/button';
import { EnumService, AccountTypeEnum } from '../../services/enum.service';
import { MultiSelectModule } from 'primeng/multiselect';
import { SelectModule } from 'primeng/select';
import { OverlayModule } from 'primeng/overlay';
import { Currency, CurrencySchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/currency_pb';
import { create } from '@bufbuild/protobuf';
import { ListTagsResponse_TagItem, TagsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/tags/v1/tags_pb';
import { ReturnUrlHelper } from '../../shared/helpers/return-url.helper';
import { TableQueryStateHelper } from '../../shared/helpers/table-query-state.helper';
import { TableStatePersistence } from '../../shared/helpers/table-state-persistence.helper';
import { TabSessionService } from '../../shared/services/tab-session.service';

@Component({
    selector: 'app-tags-list',
    templateUrl: 'tags-list.component.html',
    imports: [OverlayModule, FormsModule, InputText, ToastModule, TableModule, InputIcon, IconField, Button, MultiSelectModule, SelectModule, RouterLink],
    styles: `
        :host ::ng-deep .tagListingTable .p-datatable-header {
            border-width: 0 !important;
        }
    `
})
export class TagsListComponent implements OnInit, AfterViewInit {
    @ViewChild('dt1', { static: false }) table!: Table;

    statuses: any[] = [];

    loading: boolean = false;

    public tags: ListTagsResponse_TagItem[] = [];
    private tagsService;

    public filters: { [s: string]: FilterMetadata } = {};
    public multiSortMeta: any[] = [
        {
            field: 'tag.id',
            order: 1
        }
    ];

    @ViewChild('filter') filter!: ElementRef;
    public initialGlobalFilter: string = '';

    private readonly stateKey = 'tags';

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        public router: Router,
        route: ActivatedRoute,
        private tabSession: TabSessionService
    ) {
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
        });
    }

    getAccountUrl(account: ListTagsResponse_TagItem): string {
        return this.router.createUrlTree(['/', 'tags', account.tag!.id.toString()]).toString();
    }

    public get currentReturnUrl(): string {
        return ReturnUrlHelper.build(this.router);
    }

    async ngOnInit() {
        this.loading = true;

        try {
            let resp = await this.tagsService.listTags({});
            this.tags = resp.tags || [];
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

    protected readonly TimestampHelper = TimestampHelper;
}
