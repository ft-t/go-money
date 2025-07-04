import { Component } from '@angular/core';
import { TransactionsTableComponent } from '../../shared/components/transactions-table/transactions-table.component';
import { ActivatedRoute } from '@angular/router';
import { TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';

@Component({
    selector: 'app-transactions-list',
    imports: [TransactionsTableComponent],
    templateUrl: './transactions-list.component.html',
})
export class TransactionsListComponent {
    public newTransactionType: TransactionType | null = null;

    constructor(
        route: ActivatedRoute,
    ) {
        route.data.subscribe((data) => {
            if(data['newTransactionType']) {
                this.newTransactionType = data['newTransactionType'];
            }
        })
    }

    // Additional methods and properties can be added here as needed
}
