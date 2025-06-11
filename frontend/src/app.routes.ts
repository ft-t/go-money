import { Routes } from '@angular/router';
import { AppLayout } from './app/layout/component/app.layout';
import { Notfound } from './app/pages/notfound/notfound';
import { authGuard } from './app/services/guards/auth.guard';
import { LoginComponent } from './app/modules/auth/login/login.component';
import { AccountListComponent } from './app/modules/accounts/account-list/account-list.component';
import { AccountUpsertComponent } from './app/modules/accounts/account-list/account-upsert/account-upsert.component';

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
                component: AccountListComponent
            },
            {
                path: 'accounts/new',
                component: AccountListComponent
            },
            {
                path: 'accounts/edit/:id',
                component: AccountListComponent
            },
            {
                path: 'account/:id',
                component: AccountUpsertComponent
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
