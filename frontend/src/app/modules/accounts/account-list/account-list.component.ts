import { Component, ElementRef, Inject, OnInit, ViewChild } from '@angular/core';
import { Table, TableModule } from 'primeng/table';
import { FormsModule } from '@angular/forms';
import { InputText } from 'primeng/inputtext';
import { ToastModule } from 'primeng/toast';
import { InputIcon } from 'primeng/inputicon';
import { IconField } from 'primeng/iconfield';
import { TRANSPORT_TOKEN } from '../../../consts/transport';
import { Transport, createClient } from '@connectrpc/connect';
import { AccountsService, ListAccountsResponse_AccountItem } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { ErrorHelper } from '../../../helpers/error.helper';
import { FilterMetadata, MessageService } from 'primeng/api';
import { CommonModule, DatePipe } from '@angular/common';
import { TimestampHelper } from '../../../helpers/timestamp.helper';
import { ActivatedRoute, Router } from '@angular/router';
import { Button } from 'primeng/button';
import { EnumService, AccountTypeEnum } from '../../../services/enum.service';
import { MultiSelectModule } from 'primeng/multiselect';
import { SelectModule } from 'primeng/select';
import { OverlayModule } from 'primeng/overlay';

@Component({
    selector: 'app-account-list',
    templateUrl: 'account-list.component.html',
    imports: [OverlayModule, FormsModule, InputText, ToastModule, TableModule, InputIcon, IconField, DatePipe, Button, MultiSelectModule, SelectModule, CommonModule]
})
export class AccountListComponent implements OnInit {
    statuses: any[] = [];

    loading: boolean = false;

    public accountTypesMap: { [id: string]: AccountTypeEnum } = {};

    public accounts: ListAccountsResponse_AccountItem[] = [];
    private accountService;
    public accountTypes = EnumService.getAccountTypes();
    public filters: { [s: string]: FilterMetadata } = {};

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

    async ngOnInit() {
        this.loading = true;

        for (let type of this.accountTypes) {
            this.accountTypesMap[type.value] = type;
        }

        try {
            let resp = await this.accountService.listAccounts({});
            this.accounts = resp.accounts || [];
            console.log(this.accounts);
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
