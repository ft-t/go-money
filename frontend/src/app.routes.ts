import { Routes } from '@angular/router';
import { AppLayout } from './app/layout/component/app.layout';
import { Dashboard } from './app/pages/dashboard/dashboard';
import { Documentation } from './app/pages/documentation/documentation';
import { Landing } from './app/pages/landing/landing';
import { Notfound } from './app/pages/notfound/notfound';
import { loginGuard } from './app/services/guards/login.guard';
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
        path: 'landing',
        component: Landing
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


// import { Routes } from '@angular/router';
// import { AppLayout } from './app/layout/component/app.layout';
// import { Dashboard } from './app/pages/dashboard/dashboard';
// import { Documentation } from './app/pages/documentation/documentation';
// import { Landing } from './app/pages/landing/landing';
// import { Notfound } from './app/pages/notfound/notfound';
//
// export const appRoutes: Routes = [
//   {
//     path: '',
//     component: AppLayout,
//     children: [
//       {
//         path: '',
//         component: Dashboard
//       },
//       {
//         path: 'uikit',
//         loadChildren: () => import('./app/pages/uikit/uikit.routes')
//       },
//       {
//         path: 'documentation',
//         component: Documentation
//       },
//       {
//         path: 'pages',
//         loadChildren: () => import('./app/pages/pages.routes')
//       }
//     ]
//   },
//   {
//     path: 'landing',
//     component: Landing
//   },
//   {
//     path: 'notfound',
//     component: Notfound
//   },
//   {
//     path: 'auth',
//     loadChildren: () => import('./app/pages/auth/auth.routes')
//   },
//   {
//     path: '**',
//     redirectTo: '/notfound'
//   }
// ];
