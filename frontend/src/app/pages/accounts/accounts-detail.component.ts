import { Component, Inject, OnInit } from '@angular/core';
import { TransactionsListComponent } from '../transactions/transactions-list.component';
import { ActivatedRoute } from '@angular/router';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { AccountsService, ListAccountsRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { Account, AccountSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { create } from '@bufbuild/protobuf';
import { ErrorHelper } from '../../helpers/error.helper';
import { MessageService } from 'primeng/api';

@Component({
    selector: 'app-accounts-detail',
    imports: [TransactionsListComponent],
    templateUrl: './accounts-detail.component.html'
})
export class AccountsDetailComponent implements OnInit {
    private accountsService;

    protected currentAccountId: number | undefined = undefined;
    public currentAccount: Account = create(AccountSchema, {});

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private activeRoute: ActivatedRoute,
        private messageService: MessageService
    ) {
        console.log(activeRoute)
        this.accountsService = createClient(AccountsService, this.transport);
        this.currentAccountId = parseInt(this.activeRoute.snapshot.params['accountId']) ?? undefined;
    }

    getTableTitle(): string {
        return `Transactions for "${this.currentAccount.name}"`;
    }

    async ngOnInit() {
        try {
            let accountDetails = await this.accountsService.listAccounts(
                create(ListAccountsRequestSchema, {
                    ids: [this.currentAccountId!]
                })
            );

            this.currentAccount = accountDetails.accounts[0].account!;
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }
}
