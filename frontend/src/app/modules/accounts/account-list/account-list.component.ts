import { ChangeDetectionStrategy, Component, ElementRef, Inject, OnInit, ViewChild } from '@angular/core';
import { Table, TableModule } from 'primeng/table';
import { FormsModule } from '@angular/forms';
import { InputText } from 'primeng/inputtext';
import { ToastModule } from 'primeng/toast';
import { InputIcon } from 'primeng/inputicon';
import { IconField } from 'primeng/iconfield';
import { TRANSPORT_TOKEN } from '../../../consts/transport';
import { Transport, createClient } from '@connectrpc/connect';
import {
    AccountsService,
    ListAccountsResponse_AccountItem
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { Account } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { ErrorHelper } from '../../../helpers/error.helper';
import { MessageService } from 'primeng/api';

@Component({
    selector: 'app-account-list',
    templateUrl: 'account-list.component.html',
    imports: [FormsModule, InputText, ToastModule, TableModule, InputIcon, IconField]
})
export class AccountListComponent implements OnInit {
    statuses: any[] = [];

    loading: boolean = false;

    public accounts: ListAccountsResponse_AccountItem[] = [];
    private accountService;

    @ViewChild('filter') filter!: ElementRef;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
    ) {
        this.accountService = createClient(AccountsService, this.transport);
    }

    async ngOnInit() {
        this.loading = true;

        try {
            let resp = await this.accountService.listAccounts({});
            this.accounts = resp.accounts || [];
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.GetMessage(e) });
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
}
