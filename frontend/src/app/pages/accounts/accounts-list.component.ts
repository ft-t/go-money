import { Component, ElementRef, Inject, OnInit, ViewChild } from '@angular/core';
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
    ],
    styles: `
        :host ::ng-deep .accountListTable .p-datatable-header {
            border-width: 0 !important;
        }
    `
})
export class AccountsListComponent implements OnInit {
    @ViewChild('dt1', { static: false }) table!: Table;

    statuses: any[] = [];

    loading: boolean = false;
    public reconciliationDialogVisible = false;

    public accountTypesMap: { [id: string]: AccountTypeEnum } = {};

    public accounts: ListAccountsResponse_AccountItem[] = [];
    private accountService;
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

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        public router: Router,
        route: ActivatedRoute
    ) {
        this.accountService = createClient(AccountsService, this.transport);

        if (route.snapshot.data['filters']) {
            for (let ob of route.snapshot.data['filters']) {
                for (let [key, value] of Object.entries(ob)) {
                    this.filters[key] = value as FilterMetadata;
                }
            }
        }
    }

    getAccountUrl(account: ListAccountsResponse_AccountItem): string {
        return this.router.createUrlTree(['/', 'accounts', account.account!.id.toString()]).toString();
    }

    async ngOnInit() {
        await this.loadAccounts();
    }

    async loadAccounts() {
        this.loading = true;

        for (let type of this.accountTypes) {
            this.accountTypesMap[type.value] = type;
        }

        let foundCurrencies: { [s: string]: boolean } = {};
        this.accountCurrencies = [];

        try {
            let resp = await this.accountService.listAccounts({});
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
        console.log("Refreshing table data");

        this.table.filter('', '', '');
    }


    protected readonly TimestampHelper = TimestampHelper;
}
