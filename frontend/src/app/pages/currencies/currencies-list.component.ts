import { AfterViewInit, Component, ElementRef, Inject, OnInit, ViewChild } from '@angular/core';
import { Table, TableModule } from 'primeng/table';
import { FormsModule } from '@angular/forms';
import { InputText } from 'primeng/inputtext';
import { ToastModule } from 'primeng/toast';
import { InputIcon } from 'primeng/inputicon';
import { IconField } from 'primeng/iconfield';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { Transport, createClient } from '@connectrpc/connect';
import { ErrorHelper } from '../../helpers/error.helper';
import { FilterMetadata, MessageService, SortMeta } from 'primeng/api';
import { CommonModule, DatePipe } from '@angular/common';
import { TimestampHelper } from '../../helpers/timestamp.helper';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { Button } from 'primeng/button';
import { MultiSelectModule } from 'primeng/multiselect';
import { SelectModule } from 'primeng/select';
import { OverlayModule } from 'primeng/overlay';
import { Currency } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/currency_pb';
import { create } from '@bufbuild/protobuf';
import { Category } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/category_pb';
import { CurrencyService, GetCurrenciesRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/currency/v1/currency_pb';
import { ReturnUrlHelper } from '../../shared/helpers/return-url.helper';
import { TableQueryStateHelper } from '../../shared/helpers/table-query-state.helper';
import { TableStatePersistence } from '../../shared/helpers/table-state-persistence.helper';
import { TabSessionService } from '../../shared/services/tab-session.service';

@Component({
    selector: 'app-currencies-list',
    templateUrl: 'currencies-list.component.html',
    imports: [OverlayModule, FormsModule, InputText, ToastModule, TableModule, InputIcon, IconField, Button, MultiSelectModule, SelectModule, CommonModule, RouterLink],
    styles: `
        :host ::ng-deep .tagListingTable .p-datatable-header {
            border-width: 0 !important;
        }
    `
})
export class CurrenciesListComponent implements OnInit, AfterViewInit {
    @ViewChild('dt1', { static: false }) table!: Table;

    statuses: any[] = [];

    loading: boolean = false;

    public currencies: Currency[] = [];
    private currenciesService;
    public multiSortMeta: SortMeta[] = [
        {
            field: 'isActive',
            order: -1
        },
        {
            field: 'id',
            order: 1
        }
    ];

    public filters: { [s: string]: FilterMetadata } = {};

    @ViewChild('filter') filter!: ElementRef;
    public initialGlobalFilter: string = '';

    private readonly stateKey = 'currencies';

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        public router: Router,
        route: ActivatedRoute,
        private tabSession: TabSessionService
    ) {
        this.currenciesService = createClient(CurrencyService, this.transport);

        if (route.snapshot.data['filters']) {
            for (let ob of route.snapshot.data['filters']) {
                for (let [key, value] of Object.entries(ob)) {
                    this.filters[key] = value as FilterMetadata;
                }
            }
        }

        const stored = TableStatePersistence.read(this.stateKey, this.tabSession.id);
        if (stored) {
            if (stored.filters) this.filters = { ...this.filters, ...(stored.filters as { [s: string]: FilterMetadata }) };
            if (stored.sort && stored.sort.length > 0) this.multiSortMeta = stored.sort;
            if (stored.global) this.initialGlobalFilter = stored.global;
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

    getDetailsUrl(entity: Category): string {
        return this.router.createUrlTree(['/', 'currencies', entity.id.toString()]).toString();
    }

    public get currentReturnUrl(): string {
        return ReturnUrlHelper.build(this.router);
    }

    async ngOnInit() {
        this.loading = true;

        try {
            let resp = await this.currenciesService.getCurrencies(
                create(GetCurrenciesRequestSchema, {
                    includeDisabled: true
                })
            );

            this.currencies = resp.currencies || [];
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
