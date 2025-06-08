import { RouterModule, Routes } from '@angular/router';
import { NgModule } from '@angular/core';
import { AccountListComponent } from './account-list.component';

const routes: Routes = [
  {
    path: '',
    component: AccountListComponent
  }
]

@NgModule({
  imports: [
    RouterModule.forChild(routes)
  ],
  exports: [
    RouterModule
  ]
})
export class AccountListRoutingModule {
}
