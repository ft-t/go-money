import { ChangeDetectionStrategy, Component, ElementRef, Inject, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { OverlayModule } from 'primeng/overlay';
import { FormsModule } from '@angular/forms';
import { InputText } from 'primeng/inputtext';
import { ToastModule } from 'primeng/toast';
import { TableModule } from 'primeng/table';
import { InputIcon } from 'primeng/inputicon';
import { IconField } from 'primeng/iconfield';
import { CommonModule, DatePipe } from '@angular/common';
import { Button } from 'primeng/button';
import { MultiSelectModule } from 'primeng/multiselect';
import { SelectModule } from 'primeng/select';
import { TRANSPORT_TOKEN } from '../../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { MessageService } from 'primeng/api';
import { Router } from '@angular/router';
import { TransactionsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';

@Component({
    selector: 'app-transaction-list',
    templateUrl: 'transaction-list.component.html',
    imports: [OverlayModule, FormsModule, InputText, ToastModule, TableModule, InputIcon, IconField, DatePipe, Button, MultiSelectModule, SelectModule, CommonModule]
})
export class TransactionListComponent implements OnInit {
    private transactionsService;
    public loading = false;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        public router: Router
    ) {
        this.transactionsService = createClient(TransactionsService, this.transport);
    }

    ngOnInit() {
        // this.transactionsService.listTransactions({})
    }
}
