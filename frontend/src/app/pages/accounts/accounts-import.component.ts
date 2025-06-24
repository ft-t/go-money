import { Component, Inject } from '@angular/core';
import { Button } from 'primeng/button';
import { DropdownModule } from 'primeng/dropdown';
import { Fluid } from 'primeng/fluid';
import { InputText } from 'primeng/inputtext';
import { NgIf } from '@angular/common';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { Textarea } from 'primeng/textarea';
import { EditorModule } from 'primeng/editor';
import { Highlight, HighlightAuto } from 'ngx-highlightjs';
import { Card, CardModule } from 'primeng/card';
import { MessageService } from 'primeng/api';
import { Toast } from 'primeng/toast';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { AccountsService, CreateAccountRequestSchema, CreateAccountsBulkRequest, CreateAccountsBulkRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { ErrorHelper } from '../../helpers/error.helper';
import { create } from '@bufbuild/protobuf';
import { Step, StepList, StepPanel, StepPanels, Stepper } from 'primeng/stepper';
import { IftaLabel } from 'primeng/iftalabel';

@Component({
    selector: 'app-accounts-import',
    imports: [Button, DropdownModule, Fluid, ReactiveFormsModule, Textarea, FormsModule, EditorModule, Highlight, Card, CardModule, Toast, Stepper, Step, StepList, StepPanels, StepPanel, IftaLabel],
    templateUrl: './accounts-import.component.html'
})
export class AccountsImportComponent {
    public clientData: string = '';
    public migrationScript = `SELECT json_agg(result)
                              FROM (SELECT a.name,
                                           tc.code AS currency,
                                           CASE a.account_type_id
                                               WHEN 11 THEN 4 -- liability
                                               ELSE 1 -- regular
                                               END AS type,
                                           (SELECT n.text
                                            FROM notes n
                                            WHERE n.noteable_id = a.id
                                              AND n.noteable_type = 'FireflyIII\\Models\\Account'
                                                      LIMIT 1) AS note,
                                   (SELECT CASE
                                               WHEN n.text ~ '^\\s*[\\{\\[]'
                                THEN (SELECT jsonb_object_agg(key, value)
                                      FROM jsonb_each_text(n.text::jsonb))
                                               ELSE jsonb_build_object('note', n.text)
                                               END
                                    FROM notes n
                                    WHERE n.noteable_id = a.id
                                      AND n.noteable_type = 'FireflyIII\\Models\\Account' LIMIT 1) AS extra,
                                   a.iban,
                                   (SELECT replace(am2.data, '"', '')
                                    FROM account_meta am2
                                    WHERE am2.account_id = a.id
                                      AND am2.name = 'account_number' LIMIT 1) AS account_number FROM accounts a
                  JOIN public.account_meta am
                              ON a.id = am.account_id AND am.name = 'currency_id'
                                  JOIN public.transaction_currencies tc
                                  ON replace(am.data, '"', ''):: int = tc.id
                              WHERE a.account_type_id NOT IN (6, 13, 10, 2)
                              ORDER BY a."order" ASC
                                  ) result;
    `;

    private accountService;
    public reviewText: string = '';
    public importDisabled: boolean = true;

    constructor(
        private messageService: MessageService,
        @Inject(TRANSPORT_TOKEN) private transport: Transport
    ) {
        this.accountService = createClient(AccountsService, this.transport);
    }

    onValueChange(event: number) {
        if (event === 3) {
            this.reviewText = this.getReviewData();
        }
    }

    getReviewData() {
        try {
            let parsed = this.getRequest();

            let result = `Accounts to be imported: ${parsed.accounts.length}\n`;

            this.importDisabled = false;

            return result;
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            this.importDisabled = true;
            return ErrorHelper.getMessage(e);
        }
    }

    getRequest(): CreateAccountsBulkRequest {
        let parsedData = JSON.parse(this.clientData) as any[];

        if (!parsedData || !Array.isArray(parsedData) || parsedData.length === 0) {
            throw new Error('No valid data provided for import. Should be a json array of objects.');
        }

        let req = create(CreateAccountsBulkRequestSchema, {
            accounts: []
        });

        for (let item of parsedData) {
            req.accounts.push(
                create(CreateAccountRequestSchema, {
                    extra: item.extra,
                    name: item.name,
                    currency: item.currency,
                    type: item.type,
                    iban: item.iban,
                    accountNumber: item.account_number,
                    note: item.note
                })
            );
        }

        return req;
    }

    async import() {
        try {
            let req = this.getRequest();

            this.importDisabled = true;
            let response = await this.accountService.createAccountsBulk(req);

            let text = `Import completed successfully. Created ${response.createdCount} accounts, skipped ${response.duplicateCount} accounts.\n`;

            for (let message of response.messages) {
                text += `\n${message}`;
            }

            this.reviewText = text;

            this.messageService.add({
                severity: 'info',
                detail: `Import completed successfully. Created ${response.createdCount} accounts, skipped ${response.duplicateCount} accounts`
            });
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }
}
