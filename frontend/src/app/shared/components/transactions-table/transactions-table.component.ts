import { AfterViewInit, Component, Inject, Input, OnChanges, OnInit, SimpleChanges, ViewChild } from '@angular/core';
import { OverlayModule } from 'primeng/overlay';
import { FormsModule } from '@angular/forms';
import { ToastModule } from 'primeng/toast';
import { Table, TableLazyLoadEvent, TableModule } from 'primeng/table';
import { CommonModule, DatePipe } from '@angular/common';
import { Button } from 'primeng/button';
import { MultiSelectModule } from 'primeng/multiselect';
import { SelectModule } from 'primeng/select';
import { TRANSPORT_TOKEN } from '../../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { FilterMetadata, MessageService, SortMeta } from 'primeng/api';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { ListTransactionsRequest_SortSchema, ListTransactionsRequestSchema, SortField, TransactionsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { Transaction, TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { TimestampHelper } from '../../../helpers/timestamp.helper';
import { AccountsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { AccountTypeEnum, EnumService } from '../../../services/enum.service';
import { Account } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { ErrorHelper } from '../../../helpers/error.helper';
import { create } from '@bufbuild/protobuf';
import { SelectedDateService } from '../../../core/services/selected-date.service';
import { TimestampSchema } from '@bufbuild/protobuf/wkt';
import { BusService } from '../../../core/services/bus.service';
import { combineLatest, Observable, skip, Subscription, switchMap } from 'rxjs';
import { TagsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/tags/v1/tags_pb';
import { Tag } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/tag_pb';
import { FancyTagComponent } from '../fancy-tag/fancy-tag.component';

export class FilterWrapper {
    public filters: { [s: string]: FilterMetadata } | undefined;
}

@Component({
    selector: 'app-transaction-table',
    templateUrl: 'transactions-table.component.html',
    imports: [OverlayModule, FormsModule, ToastModule, TableModule, DatePipe, Button, MultiSelectModule, SelectModule, CommonModule, RouterLink, FancyTagComponent],
    styles: `
        :host ::ng-deep .transactionListingTable .p-datatable-header {
            border-width: 0 !important;
        }
    `
})
export class TransactionsTableComponent implements OnInit, OnChanges {
    private transactionsService;
    public loading = false;
    public transactions: Transaction[] = [];
    public accountsService;
    public tagsService;
    public transactionTypes: AccountTypeEnum[] = EnumService.getAllTransactionTypes();
    public transactionTypesMap: { [id: string]: AccountTypeEnum } = {};

    public filters: { [s: string]: FilterMetadata } = {};
    public accountMap: { [id: number]: Account } = {};
    public tagsMap: { [id: number]: Tag } = {};
    public accounts: Account[] = [];
    public tags: Tag[] = [];

    public maxSelectedLabels = 1;

    @Input() filtersWrapper: FilterWrapper | undefined;

    @Input() transactionTypeForCreate: TransactionType | null = null;

    @Input() tableTitle: string = 'Transactions';

    @Input() public currentAccountId: number | undefined;

    public ignoreDateFilter: boolean = false;
    private lastEvent: TableLazyLoadEvent | undefined;
    public totalRecords: number = 0;
    public multiSortMeta: SortMeta[] = [
        {
            field: 'transactionItem.transactionDate.nanos',
            order: -1
        }
    ];

    @ViewChild('dt1', { static: false }) table!: Table;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        public router: Router,
        private selectedDateService: SelectedDateService,
        routeSnapshot: ActivatedRoute
    ) {
        for (let type of this.transactionTypes) {
            this.transactionTypesMap[type.value] = type;
        }

        routeSnapshot.data.subscribe((data) => {
            if (data['preselectedFilter']) {
                this.filters = data['preselectedFilter'];
            }
        });

        routeSnapshot.queryParams.subscribe((params) => {
            if (params['ignoreDateFilter'] === 'true') {
                this.ignoreDateFilter = true;
            }

            if (params['title']) {
                this.filters['title'] = {
                    value: params['title'],
                    matchMode: 'contains'
                } as FilterMetadata;
            }

            this.refreshTable();
        });

        this.transactionsService = createClient(TransactionsService, this.transport);
        this.accountsService = createClient(AccountsService, this.transport);
        this.tagsService = createClient(TagsService, this.transport);
    }

    ngOnChanges(changes: SimpleChanges): void {
        if (changes['currentAccountId']) {
            this.refreshTable();
        }
    }

    async createNewTransaction() {
        await this.router.navigate(['/transactions', 'new'], {
            queryParams: {
                type: this.transactionTypeForCreate ?? TransactionType.WITHDRAWAL
            }
        });
    }

    async ngOnInit() {
        await Promise.all([this.fetchAccounts(), this.fetchTags()]);

        if (this.filtersWrapper && this.filtersWrapper.filters) {
            Object.assign(this.filters, this.filtersWrapper.filters);
        }

        this.selectedDateService.fromDate.pipe(skip(1)).subscribe(() => {
            this.refreshTable();
        });

        this.selectedDateService.toDate.pipe(skip(1)).subscribe(() => {
            this.refreshTable();
        });

        this.refreshTable();
    }

    async fetchTags() {
        this.tags = [];

        try {
            let resp = await this.tagsService.listTags({});
            for (let account of resp.tags) {
                this.tagsMap[account.tag!.id] = account.tag!;
                this.tags.push(account.tag!);
            }
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    async fetchAccounts() {
        try {
            let resp = await this.accountsService.listAccounts({});
            for (let account of resp.accounts) {
                this.accountMap[account.account!.id] = account.account!;
                this.accounts.push(account.account!);
            }
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    getFilterIcon(): string {
        if (this.ignoreDateFilter) return 'pi pi-filter';

        return 'pi pi-filter-slash';
    }

    async switchDateFilter() {
        this.ignoreDateFilter = !this.ignoreDateFilter;

        if (this.lastEvent) await this.fetchTransactions(this.lastEvent);
    }

    refreshTable() {
        if (!this.table) {
            return;
        }

        this.table.filter('', '', '');
    }

    getAccountName(accountId: number | undefined): string {
        if (!accountId) {
            return '';
        }

        let account = this.accountMap[accountId];
        if (!account) {
            return '';
        }

        return account.name || '';
    }

    getTag(tagID: number): Tag | undefined {
        return this.tagsMap[tagID];
    }

    paramsToQueryString(filters: { [s: string]: FilterMetadata }) {
        console.log('constructFilters', filters);
    }

    async fetchTransactions(event: TableLazyLoadEvent) {
        console.log(event);

        this.lastEvent = event;
        this.loading = true;

        let req = create(ListTransactionsRequestSchema, {
            limit: event.rows ?? 50,
            skip: event.first ?? 0
        });

        let fromDate = this.selectedDateService.fromDate.value;
        let toDate = this.selectedDateService.toDate.value;

        if (!this.ignoreDateFilter) {
            req.fromDate = create(TimestampSchema, {
                seconds: BigInt(Math.floor(fromDate.getTime() / 1000)),
                nanos: (fromDate.getMilliseconds() % 1000) * 1_000_000
            });

            req.toDate = create(TimestampSchema, {
                seconds: BigInt(Math.floor(toDate.getTime() / 1000)),
                nanos: (toDate.getMilliseconds() % 1000) * 1_000_000
            });
        }

        if (this.currentAccountId) {
            req.anyAccountIds = req.anyAccountIds ?? [];

            req.anyAccountIds.push(this.currentAccountId);
        }

        if (event.filters) {
            let sourceAccountIds = event.filters['sourceAccountIds'] as FilterMetadata;

            if (sourceAccountIds && sourceAccountIds.value && Array.isArray(sourceAccountIds.value)) {
                req.sourceAccountIds = sourceAccountIds.value.map((id) => parseInt(id as string));
            }

            let destinationAccountIds = event.filters['destinationAccountIds'] as FilterMetadata;
            if (destinationAccountIds && destinationAccountIds.value && Array.isArray(destinationAccountIds.value)) {
                req.destinationAccountIds = destinationAccountIds.value.map((id) => parseInt(id as string));
            }

            let title = event.filters['title'] as FilterMetadata;
            if (title && title.value && typeof title.value === 'string' && title.value.trim() !== '') {
                req.textQuery = title.value.trim();
            }

            let transactionTypes = event.filters['transactionTypes'] as FilterMetadata;
            if (transactionTypes && transactionTypes.value && Array.isArray(transactionTypes.value)) {
                req.transactionTypes = transactionTypes.value.map((type) => type as TransactionType);
            }

            let tags = event.filters['tags'] as FilterMetadata;
            if (tags && tags.value && Array.isArray(tags.value)) {
                req.tagIds = tags.value.map((id) => parseInt(id as string));
            }

            // let idFilter = event.filters['id'] as FilterMetadata
            // if(idFilter)
            //     req.tagIds
        }

        if (event.multiSortMeta) {
            for (let sortData of event.multiSortMeta) {
                let sortReq = create(ListTransactionsRequest_SortSchema, {
                    ascending: sortData.order == 1
                });

                switch (sortData.field) {
                    case 'transactionItem.transactionDate.nanos':
                        sortReq.field = SortField.TRANSACTION_DATE;
                        break;
                    default:
                        console.log('Unknown sort field:', sortData.field);
                        continue;
                }

                req.sort.push(sortReq);
            }
        }

        try {
            let resp = await this.transactionsService.listTransactions(req);

            this.transactions = resp.transactions;
            this.totalRecords = Number(resp.totalCount);
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        } finally {
            this.loading = false;
        }
    }

    getTransactionType(transaction: Transaction): string {
        if (!transaction.type) {
            return 'Unknown';
        }

        let type = this.transactionTypesMap[transaction.type];
        if (!type) {
            return 'Unknown';
        }

        return type.name;
    }

    getTransactionTypeColor(transaction: Transaction): string[] {
        return ['text-wrap', 'break-all', this.getAmountColor(transaction)];
    }

    getTransactionLink(id: number): string {
        return this.router.createUrlTree(['/', 'transactions', 'edit', id.toString()]).toString();
    }

    getAmountColor(transaction: Transaction): string {
        switch (transaction.type) {
            case TransactionType.WITHDRAWAL:
                return 'text-red-500';
            case TransactionType.DEPOSIT:
                return 'text-green-500';
            case TransactionType.TRANSFER_BETWEEN_ACCOUNTS:
                return 'text-blue-500';
            default:
                return 'text-gray-500';
        }
    }

    getAccountUrl(accountId: number | undefined): string {
        if (!accountId) {
            return '';
        }

        return this.router.createUrlTree(['/', 'accounts', accountId.toString()]).toString();
    }

    getAccountColorClass(accID: number | undefined): string[] {
        let result = ['text-wrap', 'break-all'];

        if (this.currentAccountId && this.currentAccountId === accID) {
            result.push('text-purple-500');
        }

        return result;
    }

    abs(val: string | undefined): number | undefined {
        if (val === undefined) {
            return undefined;
        }

        return Math.abs(parseFloat(val));
    }

    isSameAmount(transaction: Transaction): boolean {
        return this.abs(transaction.sourceAmount) == this.abs(transaction.destinationAmount) && transaction.sourceCurrency == transaction.destinationCurrency;
    }

    formatAmountV2(transaction: Transaction): string[] {
        let val: string[] = [];

        if (transaction.sourceAmount) {
            val.push(`${transaction.sourceAmount} ${transaction.sourceCurrency ?? ''}`);
        } else {
            val.push('');
        }

        if (transaction.destinationAmount) {
            val.push(`${transaction.destinationAmount} ${transaction.destinationCurrency ?? ''}`);
        } else {
            val.push('');
        }

        return val;
    }

    formatAmounts(transaction: Transaction): string {
        let val = '';

        let hasSource = false;

        if (transaction.sourceAmount) {
            val += `${transaction.sourceAmount} ${transaction.sourceCurrency ?? ''}`;
            hasSource = true;
        }

        if (this.abs(transaction.sourceAmount) == this.abs(transaction.destinationAmount) && transaction.sourceCurrency == transaction.destinationCurrency) {
            return val;
        }

        if (transaction.destinationAmount) {
            if (hasSource) {
                val += ' (';
            }
            val += `${transaction.destinationAmount} ${transaction.destinationCurrency ?? ''}`;

            if (hasSource) {
                val += ')';
            }
        }

        return val.trim();
    }

    protected readonly TimestampHelper = TimestampHelper;
}
