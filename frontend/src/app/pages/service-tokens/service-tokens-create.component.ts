import { Component, EventEmitter, Inject, OnInit, Output } from '@angular/core';
import { Fluid } from 'primeng/fluid';
import { InputText } from 'primeng/inputtext';
import { FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { MessageService } from 'primeng/api';
import { ErrorHelper } from '../../helpers/error.helper';
import { NgIf } from '@angular/common';
import { Button } from 'primeng/button';
import { ConfigurationService, CreateServiceTokenRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/configuration/v1/configuration_pb';
import { create } from '@bufbuild/protobuf';
import { DatePickerModule } from 'primeng/datepicker';
import { TimestampHelper } from '../../helpers/timestamp.helper';
@Component({
    selector: 'app-service-tokens-create',
    imports: [Fluid, InputText, ReactiveFormsModule, FormsModule, NgIf, Button, DatePickerModule],
    templateUrl: './service-tokens-create.component.html'
})
export class ServiceTokensCreateComponent implements OnInit {
    @Output() tokenCreated = new EventEmitter<string>();
    @Output() cancelled = new EventEmitter<void>();

    private configService;
    public form: FormGroup | undefined = undefined;
    public submitting = false;
    public minDate: Date = new Date();

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService
    ) {
        this.configService = createClient(ConfigurationService, this.transport);
    }

    ngOnInit() {
        const defaultExpiration = new Date();
        defaultExpiration.setFullYear(defaultExpiration.getFullYear() + 1);

        this.form = new FormGroup({
            name: new FormControl('', Validators.required),
            expiresAt: new FormControl(defaultExpiration, Validators.required)
        });
    }

    get name() {
        return this.form!.get('name')!;
    }

    get expiresAt() {
        return this.form!.get('expiresAt')!;
    }

    async submit() {
        this.form!.markAllAsTouched();

        if (!this.form!.valid) {
            return;
        }

        this.submitting = true;

        try {
            const expiresAtTs = TimestampHelper.dateToTimestamp(this.expiresAt.value);
            const response = await this.configService.createServiceToken(
                create(CreateServiceTokenRequestSchema, {
                    name: this.name.value,
                    expiresAt: expiresAtTs
                })
            );

            this.messageService.add({ severity: 'success', detail: 'Service token created' });
            this.tokenCreated.emit(response.token);
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        } finally {
            this.submitting = false;
        }
    }

    cancel() {
        this.cancelled.emit();
    }
}
