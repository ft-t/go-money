import { Routes } from '@angular/router';
import { AppLayout } from './app/layout/component/app.layout';
import { Notfound } from './app/pages/notfound/notfound';
import { authGuard } from './app/services/guards/auth.guard';
import { LoginComponent } from './app/modules/auth/login/login.component';
import { AccountListComponent } from './app/modules/accounts/account-list/account-list.component';
import { AccountUpsertComponent } from './app/modules/accounts/account-list/account-upsert/account-upsert.component';
import { TransactionListComponent } from './app/modules/transactions/transaction-list/transaction-list.component';

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
                path: 'accounts',
                component: AccountListComponent,
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
                component: AccountListComponent,
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
                component: AccountUpsertComponent,
                data: {
                    isEdit: false
                }
            },
            {
                path: 'accounts/edit/:id',
                component: AccountUpsertComponent,
                data: {
                    isEdit: true
                }
            },
            {
                path: 'transactions',
                component: TransactionListComponent,
                data: {
                }
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
