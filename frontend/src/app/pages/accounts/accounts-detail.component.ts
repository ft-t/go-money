import { Component, Inject, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { AccountsService, ListAccountsRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { Account, AccountSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { create } from '@bufbuild/protobuf';
import { ErrorHelper } from '../../helpers/error.helper';
import { MessageService } from 'primeng/api';
import { TransactionsTableComponent } from '../../shared/components/transactions-table/transactions-table.component';
import { BusService } from '../../core/services/bus.service';
import { BaseAutoUnsubscribeClass } from '../../objects/auto-unsubscribe/base-auto-unsubscribe-class';

@Component({
    selector: 'app-accounts-detail',
    imports: [TransactionsTableComponent],
    templateUrl: './accounts-detail.component.html'
})
export class AccountsDetailComponent extends BaseAutoUnsubscribeClass implements OnDestroy {
    private accountsService;

    public currentAccount: Account = create(AccountSchema, {});

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        activeRoute: ActivatedRoute,
        private messageService: MessageService,
        busService: BusService
    ) {
        super();

        this.accountsService = createClient(AccountsService, this.transport);

        busService.currentAccountId.subscribe(async (accountId) => {
            await this.setAccount(accountId);
        });

        activeRoute.params.subscribe((params) => {
            let parsed = parseInt(params['accountId']) ?? undefined;

            busService.currentAccountId.next(parsed);
        });
    }

    public override ngOnDestroy(): void {
        super.ngOnDestroy();
    }

    getTableTitle(): string {
        return `Transactions for "${this.currentAccount.name}"`;
    }

    async setAccount(accountID: number) {
        try {
            let accountDetails = await this.accountsService.listAccounts(
                create(ListAccountsRequestSchema, {
                    ids: [accountID]
                })
            );

            this.currentAccount = accountDetails.accounts[0].account!;
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }
}
