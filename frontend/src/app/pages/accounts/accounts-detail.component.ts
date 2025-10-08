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
import { SelectedDateService } from '../../core/services/selected-date.service';
import { AnalyticsService, GetDebitsAndCreditsSummaryRequestSchema, GetDebitsAndCreditsSummaryResponse_SummaryItem } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/analytics/v1/analytics_pb';
import { TimestampSchema } from '@bufbuild/protobuf/wkt';
import { EnumService } from '../../services/enum.service';
import { CommonModule } from '@angular/common';
import { ConfigurationService, GetConfigurationResponse, GetConfigurationResponseSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/configuration/v1/configuration_pb';

@Component({
    selector: 'app-accounts-detail',
    imports: [TransactionsTableComponent, CommonModule],
    templateUrl: './accounts-detail.component.html'
})
export class AccountsDetailComponent extends BaseAutoUnsubscribeClass implements OnDestroy {
    private accountsService;
    private analyticsService;
    private configService;

    public currentAccount: Account = create(AccountSchema, {});
    public analytics: GetDebitsAndCreditsSummaryResponse_SummaryItem | undefined = undefined;
    public accountTypeName = '';
    public serverConfig: GetConfigurationResponse = create(GetConfigurationResponseSchema, {});

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        activeRoute: ActivatedRoute,
        private messageService: MessageService,
        busService: BusService,
        private selectedDateService: SelectedDateService
    ) {
        super();

        this.accountsService = createClient(AccountsService, this.transport);
        this.analyticsService = createClient(AnalyticsService, this.transport);
        this.configService = createClient(ConfigurationService, this.transport);

        busService.currentAccountId.subscribe(async (accountId) => {
            await this.setAccount(accountId);
        });

        activeRoute.params.subscribe((params) => {
            let parsed = parseInt(params['id']) ?? undefined;

            busService.currentAccountId.next(parsed);
        });

        this.selectedDateService.fromDate.subscribe(async () => {
            await this.loadAnalytics();
        });

        this.selectedDateService.toDate.subscribe(async () => {
            await this.loadAnalytics();
        });

        this.initializeConfig();
    }

    async initializeConfig() {
        await this.loadConfig();
    }

    public override ngOnDestroy(): void {
        super.ngOnDestroy();
    }

    async loadConfig() {
        try {
            this.serverConfig = await this.configService.getConfiguration({});
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    getTableTitle(): string {
        return `Transactions for "${this.currentAccount.name}"`;
    }

    async setAccount(accountID: number) {
        if (!accountID) {
            this.currentAccount = create(AccountSchema, {});
            this.analytics = undefined;
            return;
        }

        try {
            let accountDetails = await this.accountsService.listAccounts(
                create(ListAccountsRequestSchema, {
                    ids: [accountID]
                })
            );

            this.currentAccount = accountDetails.accounts[0].account!;
            this.updateAccountTypeName();
            await this.loadAnalytics();
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    async loadAnalytics() {
        if (!this.currentAccount.id) {
            this.analytics = undefined;
            return;
        }

        try {
            const fromDate = this.selectedDateService.fromDate.value;
            const toDate = this.selectedDateService.toDate.value;

            const response = await this.analyticsService.getDebitsAndCreditsSummary(
                create(GetDebitsAndCreditsSummaryRequestSchema, {
                    accountIds: [this.currentAccount.id],
                    startAt: create(TimestampSchema, {
                        seconds: BigInt(Math.floor(fromDate.getTime() / 1000)),
                        nanos: (fromDate.getMilliseconds() % 1000) * 1_000_000
                    }),
                    endAt: create(TimestampSchema, {
                        seconds: BigInt(Math.floor(toDate.getTime() / 1000)),
                        nanos: (toDate.getMilliseconds() % 1000) * 1_000_000
                    })
                })
            );

            this.analytics = response.items[this.currentAccount.id];
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    updateAccountTypeName() {
        const accountTypes = EnumService.getAccountTypes();
        const accountType = accountTypes.find(t => t.value === this.currentAccount.type);
        this.accountTypeName = accountType?.name || 'Unknown';
    }

    formatAmount(amount: number): string {
        return parseFloat(amount.toString()).toFixed(2);
    }
}
