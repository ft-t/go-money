import { Routes } from '@angular/router';
import { AppLayout } from './app/layout/component/app.layout';
import { Notfound } from './app/pages/notfound/notfound';
import { authGuard } from './app/services/guards/auth.guard';
import { LoginComponent } from './app/modules/auth/login/login.component';

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
                loadChildren: () => import('./app/modules/accounts/account-list/account-list.module').then((m) => m.AccountListModule)
            },
            {
                path: 'account/:id',
                loadChildren: () => import('./app/modules/accounts/account-list/account-upsert/account-upsert.module').then((m) => m.AccountUpsertModule)
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
