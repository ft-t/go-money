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
import {
    ImportTagsRequest,
    ImportTagsRequestSchema,
    TagsService,
    UpdateTagRequestSchema
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/tags/v1/tags_pb';

@Component({
    selector: 'app-tags-import',
    imports: [Button, DropdownModule, Fluid, ReactiveFormsModule, Textarea, FormsModule, EditorModule, Highlight, Card, CardModule, Toast, Stepper, Step, StepList, StepPanels, StepPanel],
    templateUrl: './tags-import.component.html'
})
export class TagsImportComponent {
    public clientData: string = '';
    public migrationScript = `select json_agg(name order by name) as names
                              from (select distinct name
                                  from (select tag as name
                                  from tags
                                  where deleted_at is null
                                  union
                                  select name
                                  from budgets
                                  union
                                  select name
                                  from bills) as cntt) as distinct_names;
    `;

    private tagsService;
    public reviewText: string = '';
    public importDisabled: boolean = true;

    constructor(
        private messageService: MessageService,
        @Inject(TRANSPORT_TOKEN) private transport: Transport
    ) {
        this.tagsService = createClient(TagsService, this.transport);
    }

    onValueChange(event: number) {
        if (event === 3) {
            this.reviewText = this.getReviewData();
        }
    }

    getReviewData() {
        try {
            let parsed = this.getRequest();

            let result = `Tags to be imported: ${parsed.tags.length}\n`;

            this.importDisabled = false;

            return result;
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            this.importDisabled = true;
            return ErrorHelper.getMessage(e);
        }
    }

    getRequest(): ImportTagsRequest {
        let parsedData = JSON.parse(this.clientData) as any[];

        if (!parsedData || !Array.isArray(parsedData) || parsedData.length === 0) {
            throw new Error('No valid data provided for import. Should be a json array of objects.');
        }

        let req = create(ImportTagsRequestSchema, {});

        for (let item of parsedData) {
            req.tags.push(
                create(UpdateTagRequestSchema, {
                    name: item
                })
            );
        }

        return req;
    }

    async import() {
        try {
            let req = this.getRequest();

            this.importDisabled = true;
            let response = await this.tagsService.importTags(req);

            let text = `Import completed successfully. Created ${response.createdCount} tags, updated ${response.updatedCount} tags.\n`;

            for (let message of response.messages) {
                text += `\n${message}`;
            }

            this.reviewText = text;

            this.messageService.add({
                severity: 'info',
                detail: `Import completed successfully. Created ${response.createdCount} accounts, updated ${response.updatedCount} tags`
            });
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }
}
