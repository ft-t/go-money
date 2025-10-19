import { Component, Input, OnChanges, SimpleChanges } from '@angular/core';
import { Transaction, TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { CommonModule } from '@angular/common';

export interface TransactionSummary {
    totalExpenses: number;
    totalIncome: number;
    totalTransfers: number;
    expenseCount: number;
    incomeCount: number;
    transferCount: number;
}

@Component({
    selector: 'app-transaction-summary',
    templateUrl: './transaction-summary.component.html',
    imports: [CommonModule],
    standalone: true
})
export class TransactionSummaryComponent implements OnChanges {
    @Input() transactions: Transaction[] = [];
    @Input() loading = false;

    public summary: TransactionSummary = {
        totalExpenses: 0,
        totalIncome: 0,
        totalTransfers: 0,
        expenseCount: 0,
        incomeCount: 0,
        transferCount: 0
    };

    ngOnChanges(changes: SimpleChanges): void {
        if (changes['transactions'] || changes['loading']) {
            this.calculateSummary();
        }
    }

    calculateSummary() {
        this.summary = {
            totalExpenses: 0,
            totalIncome: 0,
            totalTransfers: 0,
            expenseCount: 0,
            incomeCount: 0,
            transferCount: 0
        };

        for (let tx of this.transactions) {
            const amount = Math.abs(parseFloat(tx.sourceAmount || '0'));

            switch (tx.type) {
                case TransactionType.EXPENSE:
                    this.summary.totalExpenses += amount;
                    this.summary.expenseCount++;
                    break;
                case TransactionType.INCOME:
                    this.summary.totalIncome += amount;
                    this.summary.incomeCount++;
                    break;
                case TransactionType.TRANSFER_BETWEEN_ACCOUNTS:
                    this.summary.totalTransfers += amount;
                    this.summary.transferCount++;
                    break;
            }
        }
    }

    formatAmount(amount: number): string {
        return amount.toFixed(2);
    }
}
