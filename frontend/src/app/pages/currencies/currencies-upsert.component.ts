import { Component, Inject, OnInit } from '@angular/core';
import { Fluid } from 'primeng/fluid';
import { InputText } from 'primeng/inputtext';
import { FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { MessageService } from 'primeng/api';
import { create } from '@bufbuild/protobuf';
import { ErrorHelper } from '../../helpers/error.helper';
import { ActivatedRoute, Router } from '@angular/router';
import { NgIf } from '@angular/common';
import { Button } from 'primeng/button';
import { Toast } from 'primeng/toast';
import { color } from 'chart.js/helpers';
import { ColorPickerModule } from 'primeng/colorpicker';
import { Message } from 'primeng/message';
import { CacheService } from '../../core/services/cache.service';
import { Currency, CurrencySchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/currency_pb';
import { CreateCurrencyRequestSchema, CurrencyService, UpdateCurrencyRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/currency/v1/currency_pb';
import { Checkbox } from 'primeng/checkbox';

@Component({
    selector: 'app-currencies-upsert',
    imports: [Fluid, InputText, ReactiveFormsModule, FormsModule, NgIf, Button, Toast, ColorPickerModule, Message, Checkbox],
    templateUrl: './currencies-upsert.component.html'
})
export class CurrenciesUpsertComponent implements OnInit {
    public currency: Currency = create(CurrencySchema, {});
    private currenciesService;
    public form: FormGroup | undefined = undefined;
    public isCreate: boolean = false;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        routeSnapshot: ActivatedRoute,
        private router: Router,
        private cache: CacheService
    ) {
        this.currenciesService = createClient(CurrencyService, this.transport);

        try {
            this.currency.id = routeSnapshot.snapshot.params['id'];
        } catch (e) {
            this.currency.id = '';
        }

        if (routeSnapshot.snapshot.data['isCreate']) {
            this.isCreate = routeSnapshot.snapshot.data['isCreate'];
        }
    }

    async ngOnInit() {
        if (!this.isCreate) {
            try {
                let response = await this.currenciesService.getCurrencies({
                    ids: [this.currency.id],
                    includeDisabled: true
                });
                if (response.currencies && response.currencies.length == 0) {
                    this.messageService.add({ severity: 'error', detail: 'currency not found' });
                    return;
                }

                this.currency = response.currencies[0] ?? create(CurrencySchema, {});
            } catch (e) {
                this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            }
        }

        this.form = new FormGroup({
            id: new FormControl(this.currency.id, Validators.required),
            rate: new FormControl(this.currency.rate, Validators.min(0.000001)),
            isActive: new FormControl(this.currency.isActive, Validators.required),
            decimalPlaces: new FormControl(this.currency.decimalPlaces, Validators.min(1))
        });
    }

    async submit() {
        this.form!.markAllAsTouched();

        this.currency = this.form!.value as Currency;
        if (!this.form!.valid) {
            return;
        }

        this.cache.clear(); // categories

        if (!this.isCreate) {
            await this.update();
        } else {
            await this.create();
        }
    }

    get id() {
        return this.form!.get('id')!;
    }

    get rate() {
        return this.form!.get('rate')!;
    }

    get decimalPlaces() {
        return this.form!.get('decimalPlaces')!;
    }

    async update() {
        try {
            let response = await this.currenciesService.updateCurrency(
                create(UpdateCurrencyRequestSchema, {
                    id: this.currency.id,
                    rate: this.currency.rate.toString(),
                    isActive: this.currency.isActive,
                    decimalPlaces: this.currency.decimalPlaces
                })
            );

            this.isCreate = false;
            this.messageService.add({ severity: 'info', detail: 'Currency updated' });
            await this.router.navigate(['/', 'currencies', response.currency!.id.toString()]);
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    async create() {
        try {
            this.currency.rate = this.currency.rate.toString();
            let response = await this.currenciesService.createCurrency(
                create(CreateCurrencyRequestSchema, {
                    currency: this.currency
                })
            );

            this.messageService.add({ severity: 'info', detail: 'Currency created' });
            await this.router.navigate(['/', 'currencies', response.currency!.id.toString()]);
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    protected readonly color = color;
}
