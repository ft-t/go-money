import { Component } from '@angular/core';
import { TransactionsListComponent } from '../transactions/transactions-list.component';

@Component({
    selector: 'app-accounts-detail',
    imports: [TransactionsListComponent],
    templateUrl: './accounts-detail.component.html'
})
export class AccountsDetailComponent {}
