import { Routes } from '@angular/router';
import { AppLayout } from './app/layout/component/app.layout';
import { Notfound } from './app/pages/notfound/notfound';
import { authGuard } from './app/services/guards/auth.guard';
import { LoginComponent } from './app/modules/auth/login/login.component';
import { TagsListComponent } from './app/pages/tags/tags-list.component';
import { AccountsUpsertComponent } from './app/pages/accounts/accounts-upsert.component';
import { TransactionUpsertComponent } from './app/pages/transactions/transactions-upsert.component';
import { AccountsImportComponent } from './app/pages/accounts/accounts-import.component';
import { TransactionsImportComponent } from './app/pages/transactions/transactions-import.component';
import { AccountsDetailComponent } from './app/pages/accounts/accounts-detail.component';
import { FilterMetadata } from 'primeng/api';
import { TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { TransactionsListComponent } from './app/pages/transactions/transactions-list.component';
import { TagsImportComponent } from './app/pages/tags/tags-import.component';
import { AccountsListComponent } from './app/pages/accounts/accounts-list.component';
import { TagsUpsertComponent } from './app/pages/tags/tags-upsert.component';
import { TagsDetailComponent } from './app/pages/tags/tags-detail.component';
import { RuleListComponent } from './app/pages/rules/rule-list.component';
import { RulesUpsertComponent } from './app/pages/rules/rules-upsert.component';
import { DashboardComponent } from './app/pages/dashboard/dashboard.component';
import { TransactionsDetailsComponent } from './app/pages/transactions/transactions-details.component';

export const appRoutes: Routes = [
    {
        path: 'login',
        component: LoginComponent
    },
    {
        path: '',
        component: AppLayout,
        canActivate: [authGuard],
        children: [
            {
              path: '',
              component: DashboardComponent,
            },
            {
                path: 'accounts',
                component: AccountsListComponent,
                data: {
                    filters: [
                        {
                            'account.type': {
                                matchMode: 'in',
                                value: [1, 2, 3]
                            }
                        }
                    ]
                }
            },
            {
                path: 'accounts/liabilities',
                component: AccountsListComponent,
                data: {
                    filters: [
                        {
                            'account.type': {
                                matchMode: 'in',
                                value: [4]
                            }
                        }
                    ]
                }
            },
            {
                path: 'accounts/new',
                component: AccountsUpsertComponent,
                data: {
                    isEdit: false
                }
            },
            {
                path: 'accounts/import',
                component: AccountsImportComponent
            },
            {
                path: 'accounts/edit/:id',
                component: AccountsUpsertComponent,
                data: {
                    isEdit: true
                }
            },
            {
                path: 'accounts/:accountId',
                component: AccountsDetailComponent,
                data: {}
            },
            {
                path: 'transactions',
                component: TransactionsListComponent,
                data: {}
            },
            {
                path: 'transactions/deposits',
                component: TransactionsListComponent,
                data: {
                    preselectedFilter: {
                        transactionTypes: {
                            matchMode: 'in',
                            value: [TransactionType.DEPOSIT]
                        }
                    },
                    newTransactionType: TransactionType.DEPOSIT
                }
            },
            {
                path: 'transactions/withdrawals',
                component: TransactionsListComponent,
                data: {
                    preselectedFilter: {
                        transactionTypes: {
                            matchMode: 'in',
                            value: [TransactionType.WITHDRAWAL]
                        }
                    },
                    newTransactionType: TransactionType.WITHDRAWAL
                }
            },
            {
                path: 'transactions/transfers',
                component: TransactionsListComponent,
                data: {
                    preselectedFilter: {
                        transactionTypes: {
                            matchMode: 'in',
                            value: [TransactionType.TRANSFER_BETWEEN_ACCOUNTS, TransactionType.RECONCILIATION]
                        }
                    },
                    newTransactionType: TransactionType.TRANSFER_BETWEEN_ACCOUNTS
                }
            },
            {
                path: 'transactions/new',
                component: TransactionUpsertComponent,
                data: {}
            },
            {
                path: 'transactions/import',
                component: TransactionsImportComponent,
                data: {}
            },
            {
                path: 'transactions/:id',
                component: TransactionsDetailsComponent,
                data: {}
            },
            {
                path: 'transactions/edit/:id',
                component: TransactionUpsertComponent,
                data: {}
            },
            {
                path: 'tags/import',
                component: TagsImportComponent,
                data: {}
            },
            {
                path: 'tags/edit/:id',
                component: TagsUpsertComponent,
                data: {}
            },
            {
                path: 'tags/:id',
                component: TagsDetailComponent,
                data: {}
            },
            {
                path: 'tags/new',
                component: TagsUpsertComponent,
                data: {}
            },
            {
                path: 'tags',
                component: TagsListComponent,
                data: {}
            },
            {
                path: 'rules',
                component: RuleListComponent,
                data: {}
            },
            {
                path: 'rules/edit/:id',
                component: RulesUpsertComponent,
                data: {}
            },
            {
                path: 'rules/new',
                component: RulesUpsertComponent,
                data: {}
            },
        ]
    },

    {
        path: 'notfound',
        component: Notfound
    },
    {
        path: '**',
        redirectTo: '/notfound'
    }
];
