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
import { CategoriesListComponent } from './app/pages/categories/categories-list.component';
import { CategoriesUpsertComponent } from './app/pages/categories/categories-upsert.component';
import { CategoriesDetailComponent } from './app/pages/categories/categories-detail.component';
import { CurrenciesListComponent } from './app/pages/currencies/currencies-list.component';
import { CurrenciesUpsertComponent } from './app/pages/currencies/currencies-upsert.component';
import { SchedulesListComponent } from './app/pages/rules/schedules-list.component';
import { SchedulesUpsertComponent } from './app/pages/rules/schedules-upsert.component';
import { MaintenanceComponent } from './app/pages/maintenance/maintenance.component';
import { ServiceTokensListComponent } from './app/pages/service-tokens/service-tokens-list.component';

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
                component: DashboardComponent
            },
            {
                path: 'accounts',
                component: AccountsListComponent,
                data: {
                    filters: [
                        {
                            'account.type': {
                                matchMode: 'in',
                                value: [1]
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
                path: 'accounts/income',
                component: AccountsListComponent,
                data: {
                    filters: [
                        {
                            'account.type': {
                                matchMode: 'in',
                                value: [6]
                            }
                        }
                    ]
                }
            },
            {
                path: 'accounts/expense',
                component: AccountsListComponent,
                data: {
                    filters: [
                        {
                            'account.type': {
                                matchMode: 'in',
                                value: [5]
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
                path: 'accounts/:id',
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
                            value: [TransactionType.INCOME]
                        }
                    },
                    newTransactionType: TransactionType.INCOME
                }
            },
            {
                path: 'transactions/withdrawals',
                component: TransactionsListComponent,
                data: {
                    preselectedFilter: {
                        transactionTypes: {
                            matchMode: 'in',
                            value: [TransactionType.EXPENSE]
                        }
                    },
                    newTransactionType: TransactionType.EXPENSE
                }
            },
            {
                path: 'transactions/transfers',
                component: TransactionsListComponent,
                data: {
                    preselectedFilter: {
                        transactionTypes: {
                            matchMode: 'in',
                            value: [TransactionType.TRANSFER_BETWEEN_ACCOUNTS, TransactionType.ADJUSTMENT]
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
                path: 'tags/new',
                component: TagsUpsertComponent,
                data: {}
            },
            {
                path: 'tags/:id',
                component: TagsDetailComponent,
                data: {}
            },
            {
                path: 'tags',
                component: TagsListComponent,
                data: {}
            },
            {
                path: 'categories',
                component: CategoriesListComponent,
                data: {}
            },
            {
                path: 'categories/edit/:id',
                component: CategoriesUpsertComponent,
                data: {}
            },
            {
                path: 'categories/new',
                component: CategoriesUpsertComponent,
                data: {}
            },
            {
                path: 'categories/:id',
                component: CategoriesDetailComponent,
                data: {}
            },

            {
                path: 'currencies',
                component: CurrenciesListComponent,
                data: {}
            },
            {
                path: 'currencies/edit/:id',
                component: CurrenciesUpsertComponent,
                data: {}
            },
            {
                path: 'currencies/new',
                component: CurrenciesUpsertComponent,
                data: {
                    isCreate: true
                }
            },
            {
                path: 'currencies/:id',
                component: CurrenciesUpsertComponent,
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

            {
                path: 'schedules',
                component: SchedulesListComponent,
                data: {}
            },
            {
                path: 'schedules/edit/:id',
                component: SchedulesUpsertComponent,
                data: {}
            },
            {
                path: 'schedules/new',
                component: SchedulesUpsertComponent,
                data: {}
            },
            {
                path: 'service-tokens',
                component: ServiceTokensListComponent,
                data: {}
            },
            {
                path: 'maintenance',
                component: MaintenanceComponent,
                data: {}
            }
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
