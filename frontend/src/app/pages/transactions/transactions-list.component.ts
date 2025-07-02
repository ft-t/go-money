import { Component } from '@angular/core';
import { TransactionsTableComponent } from '../../shared/components/transactions-table/transactions-table.component';

@Component({
    selector: 'app-transactions-list',
    imports: [TransactionsTableComponent],
    templateUrl: './transactions-list.component.html',
})
export class TransactionsListComponent {
    constructor() { }

    // Additional methods and properties can be added here as needed
}
