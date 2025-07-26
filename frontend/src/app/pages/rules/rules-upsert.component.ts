import { Component, Inject, OnInit, ViewChild } from '@angular/core';
import { Fluid } from 'primeng/fluid';
import { InputText } from 'primeng/inputtext';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
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
import { Rule, RuleSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/rule_pb';
import { CreateRuleRequestSchema, DryRunRuleRequestSchema, RulesService, UpdateRuleRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/rules/v1/rules_pb';
import { Checkbox } from 'primeng/checkbox';
import { DiffEditorComponent, DiffEditorModel, EditorComponent } from 'ngx-monaco-editor-v2';
import { editor } from 'monaco-editor';
import IStandaloneCodeEditor = editor.IStandaloneCodeEditor;
import { TextareaModule } from 'primeng/textarea';
import { InputNumberModule } from 'primeng/inputnumber';
import { TransactionsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { ScriptEditorComponent } from '../../shared/components/script-editor/script-editor.component';

@Component({
    selector: 'app-rules-upsert',
    imports: [InputNumberModule, TextareaModule, Fluid, InputText, ReactiveFormsModule, FormsModule, NgIf, Button, Toast, ColorPickerModule, Checkbox, EditorComponent, DiffEditorComponent, ScriptEditorComponent],
    templateUrl: './rules-upsert.component.html'
})
export class RulesUpsertComponent implements OnInit {
    public rule: Rule = create(RuleSchema, {});
    private rulesService;

    @ViewChild('scriptEditorComponent') scriptEditor!: ScriptEditorComponent;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        routeSnapshot: ActivatedRoute,
        private router: Router
    ) {
        this.rulesService = createClient(RulesService, this.transport);

        try {
            this.rule.id = +routeSnapshot.snapshot.params['id'];
        } catch (e) {
            this.rule.id = 0;
        }
    }

    async ngOnInit() {
        if (this.rule.id) {
            try {
                let response = await this.rulesService.listRules({ ids: [+this.rule.id] });
                if (response.rules && response.rules.length == 0) {
                    this.messageService.add({ severity: 'error', detail: 'tag not found' });
                    return;
                }

                this.rule = response.rules[0] ?? create(RuleSchema, {});
            } catch (e) {
                this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            }
        }
    }

    async update() {
        this.rule.script = await this.scriptEditor.dryRun();

        try {
            let response = await this.rulesService.updateRule(
                create(UpdateRuleRequestSchema, {
                    rule: this.rule
                })
            );

            this.messageService.add({ severity: 'info', detail: 'Rule updated' });
            await this.router.navigate(['/', 'rules']);
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    async create() {
        this.rule.script = await this.scriptEditor.dryRun();

        try {
            let response = await this.rulesService.createRule(
                create(CreateRuleRequestSchema, {
                    rule: this.rule
                })
            );

            this.messageService.add({ severity: 'info', detail: 'Rule created' });
            await this.router.navigate(['/', 'rules']);
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    protected readonly color = color;
}
