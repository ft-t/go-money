import { RouterModule, Routes } from '@angular/router';
import { NgModule } from '@angular/core';
import { AccountUpsertComponent } from './account-upsert.component';

const routes: Routes = [
  {
    path: '',
    component: AccountUpsertComponent
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
export class AccountUpsertRoutingModule {
}
