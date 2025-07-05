import { Component, Inject, OnInit } from '@angular/core';
import { Button } from 'primeng/button';
import { InputText } from 'primeng/inputtext';
import { Fluid } from 'primeng/fluid';
import { DropdownModule } from 'primeng/dropdown';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { CurrencyService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/currency/v1/currency_pb';
import { ErrorHelper } from '../../helpers/error.helper';
import { MessageService } from 'primeng/api';
import { Account, AccountSchema, AccountType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { Currency } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/currency_pb';
import { create } from '@bufbuild/protobuf';
import { EnumService } from '../../services/enum.service';
import { NgIf } from '@angular/common';
import { Textarea } from 'primeng/textarea';
import { AccountsService, CreateAccountRequestSchema, UpdateAccountRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { ActivatedRoute, Router } from '@angular/router';
import { Message } from 'primeng/message';

@Component({
    selector: 'app-account-upsert',
    templateUrl: 'accounts-upsert.component.html',
    imports: [Button, InputText, Fluid, DropdownModule, FormsModule, NgIf, Textarea, Message, ReactiveFormsModule]
})
export class AccountsUpsertComponent implements OnInit {
    private currencyService;
    private accountsService;

    public currencies: Currency[] = [];

    public account: Account = create(AccountSchema, {});
    public accountTypes = EnumService.getAccountTypes();
    public isEdit: boolean = false;
    public form: FormGroup = new FormGroup({});
    public isFormReady: boolean = false;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        private router: Router,
        private routeSnapshot: ActivatedRoute
    ) {
        this.isEdit = routeSnapshot.snapshot.data['isEdit'];

        this.currencyService = createClient(CurrencyService, this.transport);
        this.accountsService = createClient(AccountsService, this.transport);
    }

    async loadCurrencies() {
        try {
            let response = await this.currencyService.getCurrencies({});

            this.currencies = response.currencies || [];
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    async ngOnInit() {
        this.account = create(AccountSchema, {});

        await this.loadCurrencies();

        if (this.isEdit) {
            const accountId = +this.routeSnapshot.snapshot.params['id'];

            if (isNaN(accountId) || accountId <= 0) {
                this.messageService.add({ severity: 'error', detail: 'invalid account id' });
                return;
            }

            try {
                let response = await this.accountsService.listAccounts({ ids: [+accountId] });
                if (response.accounts && response.accounts.length == 0) {
                    this.messageService.add({ severity: 'error', detail: 'account not found' });
                    return;
                }

                this.account = response.accounts[0].account ?? create(AccountSchema, {});
            } catch (e) {
                this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            }
        }

        this.form = new FormGroup({
            id: new FormControl(this.account.id),
            name: new FormControl(this.account.name, Validators.required),
            type: new FormControl(this.account.type, Validators.min(1)),
            currency: new FormControl(this.account.currency, Validators.required),
            note: new FormControl(this.account.note),
            iban: new FormControl(this.account.iban),
            accountNumber: new FormControl(this.account.accountNumber),
            liabilityPercent: new FormControl(this.account.liabilityPercent),
            displayOrder: new FormControl(this.account.displayOrder),
            extra: new FormControl(this.account.extra ?? {})
        });

        this.isFormReady = true;
    }

    get name() {
        return this.form.get('name')!;
    }

    get type() {
        return this.form.get('type')!;
    }

    get currency() {
        return this.form.get('currency')!;
    }

    async submit() {
        this.form.markAllAsTouched();

        this.account = this.form.value as Account;
        if (!this.form.valid) {
            return;
        }

        if (this.isEdit) {
            await this.update();
        } else {
            await this.create();
        }
    }

    async update() {
        try {
            let response = await this.accountsService.updateAccount(
                create(UpdateAccountRequestSchema, {
                    id: this.account.id,
                    type: this.account.type,
                    name: this.account.name,
                    accountNumber: this.account.accountNumber,
                    iban: this.account.iban,
                    note: this.account.note,
                    liabilityPercent: this.account.liabilityPercent,
                    displayOrder: this.account.displayOrder,
                    extra: this.account.extra ?? {},
                })
            );

            this.messageService.add({ severity: 'info', detail: 'Account updated' });
            await this.router.navigate(['/', 'accounts', response.account!.id.toString()]);
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    async create() {
        try {
            let response = await this.accountsService.createAccount(
                create(CreateAccountRequestSchema, {
                    type: this.account.type,
                    name: this.account.name,
                    currency: this.account.currency,
                    accountNumber: this.account.accountNumber,
                    iban: this.account.iban,
                    note: this.account.note,
                    liabilityPercent: this.account.liabilityPercent,
                    displayOrder: this.account.displayOrder,
                    extra: this.account.extra ?? {}
                })
            );

            this.messageService.add({ severity: 'info', detail: 'New account created' });
            await this.router.navigate(['/', 'accounts', response.account!.id.toString()]);
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    protected readonly AccountType = AccountType;
}
