import { Component, EventEmitter, Inject, Input, OnChanges, OnInit, Output, SimpleChanges } from '@angular/core';
import { ButtonModule } from 'primeng/button';
import { InputGroupModule } from 'primeng/inputgroup';
import { InputGroupAddonModule } from 'primeng/inputgroupaddon';
import { DialogModule } from 'primeng/dialog';
import { Account } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { NgIf } from '@angular/common';
import { InputNumber } from 'primeng/inputnumber';
import { FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { ErrorHelper } from '../../../../helpers/error.helper';
import { TRANSPORT_TOKEN } from '../../../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { MessageService } from 'primeng/api';
import {
    CreateTransactionRequestSchema,
    ReconciliationSchema,
    TransactionsService
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { TimestampSchema } from '@bufbuild/protobuf/wkt';
import { create } from '@bufbuild/protobuf';
import { InputText } from 'primeng/inputtext';
import { Tag } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/tag_pb';

@Component({
    selector: 'app-reconciliation-modal',
    imports: [ButtonModule, InputGroupModule, InputGroupAddonModule, DialogModule, NgIf, InputNumber, FormsModule, ReactiveFormsModule, InputText],
    templateUrl: './reconciliation-modal.component.html'
})
export class ReconciliationModalComponent implements OnChanges {
    private transactionService;
    public transactionDate: Date = new Date();

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService
    ) {
        this.transactionService = createClient(TransactionsService, this.transport);
    }

    @Input()
    public visible = false;

    @Input()
    public account: Account | undefined = undefined;

    @Output()
    closed = new EventEmitter<boolean>(); // for two-way binding

    public form: FormGroup | undefined = undefined;

    ngOnChanges(changes: SimpleChanges): void {
        if (changes['account']) {
            this.setForm();
        }
    }

    setForm() {
        let title = `Account reconciliation at ${this.transactionDate.toLocaleDateString()} ${this.transactionDate.toLocaleTimeString()}`;
        this.form = new FormGroup({
            currentBalance: new FormControl(parseFloat(this.account?.currentBalance || '0')),
            newBalance: new FormControl(undefined, Validators.required),
            difference: new FormControl(undefined),
            title: new FormControl(title, Validators.required)
        });

        this.form.get('currentBalance')!.disable();
        this.form.get('difference')!.disable();

        this.form.get('newBalance')!.valueChanges.subscribe((newVal) => {
            const current = parseFloat(this.account?.currentBalance || '0');
            const diff = newVal !== null && !isNaN(newVal) ? newVal - current : null;

            this.form!.get('difference')!.setValue(diff, { emitEvent: false });
        });
    }

    get currentBalance() {
        return this.form?.get('currentBalance')?.value;
    }

    onHide() {
        this.closed.emit(true);
        this.visible = false;
    }

    async save() {
        this.form!.markAllAsTouched();

        if (!this.form!.valid) {
            return;
        }

        try {
            let baseRequest = create(CreateTransactionRequestSchema, {
                notes: this.form!.get('title')!.value,
                extra: {},
                tagIds: [],
                transactionDate: create(TimestampSchema, {
                    seconds: BigInt(Math.floor(this.transactionDate.getTime() / 1000)),
                    nanos: (this.transactionDate.getMilliseconds() % 1000) * 1_000_000
                }),
                title: this.form!.get('title')!.value
            });

            baseRequest.transaction.case = 'reconciliation';
            baseRequest.transaction.value = create(ReconciliationSchema, {
                destinationAccountId: this.account!.id,
                destinationAmount: this.form!.get('difference')!.value.toString(),
                destinationCurrency: this.account!.currency
            });

            let resp = await this.transactionService.createTransaction(baseRequest);

            this.messageService.add({
                severity: 'info',
                detail: `Reconciliation transaction created. ID: ${resp.transaction!.id}`
            });

            this.onHide();
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }
}
