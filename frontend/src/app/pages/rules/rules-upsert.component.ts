import { Component, Inject, OnInit } from '@angular/core';
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

@Component({
    selector: 'app-rules-upsert',
    imports: [InputNumberModule, TextareaModule, Fluid, InputText, ReactiveFormsModule, FormsModule, NgIf, Button, Toast, ColorPickerModule, Checkbox, EditorComponent, DiffEditorComponent],
    templateUrl: './rules-upsert.component.html'
})
export class RulesUpsertComponent implements OnInit {
    public rule: Rule = create(RuleSchema, {});
    private rulesService;
    private transactionService;

    editorOptions = { theme: 'vs-dark', language: 'lua' };
    diffOptions = {
        theme: 'vs-dark'
    };
    originalModel: DiffEditorModel = {
        code: '',
        language: 'text/json'
    };

    public modifiedModel: DiffEditorModel = {
        code: '',
        language: 'text/json'
    };

    public helpContent = this.generateSimpleApiHelp(this.getSuggestions());
    public dryRunTransactionId: number = 0;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        routeSnapshot: ActivatedRoute,
        private router: Router
    ) {
        this.rulesService = createClient(RulesService, this.transport);
        this.transactionService = createClient(TransactionsService, this.transport);

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

    originalTextModel: editor.ITextModel | null = null;
    modifiedTextModel: editor.ITextModel | null = null;

    onDiffEditorInit(diffEditor: any) {
        const model = diffEditor.getModel();
        this.originalTextModel = model.original;
        this.modifiedTextModel = model.modified;
    }

    async dryRun() {
        await this.ensureTxSet();

        try {
            let response = await this.rulesService.dryRunRule(
                create(DryRunRuleRequestSchema, {
                    rule: this.rule,
                    transactionId: BigInt(+this.dryRunTransactionId)
                })
            );

            this.originalTextModel!.setValue(JSON.stringify(response.before, null, 2));

            this.modifiedTextModel!.setValue(JSON.stringify(response.after, null, 2));
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            throw e;
        }
    }

    async update() {
        await this.dryRun();

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
        await this.dryRun();
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

    generateSimpleApiHelp(suggestions: any[]) {
        const lines = ['Transaction API:', ''];

        const added = new Set();

        for (const s of suggestions) {
            if (added.has(s.label)) continue;
            lines.push(`${s.label} — ${s.documentation}`);
            added.add(s.label);
        }

        return lines.join('\n');
    }

    async ensureTxSet() {
        if (this.dryRunTransactionId) return;

        try {
            let txs = await this.transactionService.listTransactions({
                limit: 1
            });

            if (txs.transactions.length == 0) {
                this.messageService.add({
                    severity: 'error',
                    detail: 'No transactions found for dry run. Please create at least 1 transaction'
                });
                return;
            }

            this.dryRunTransactionId = Number(txs.transactions[0].id);
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    getSuggestions() {
        let monaco = (window as any).monaco;

        let kind = 1;
        let snippet = 4;

        if (monaco) {
            kind = monaco.languages.CompletionItemKind.Function;
            snippet = monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet;
        }

        let suggestions = [];

        suggestions.push({
            label: `helpers:getAccountByID(value)`,
            kind: kind,
            insertText: `helpers:getAccountByID(value)`,
            insertTextRules: snippet,
            documentation: `Get account by ID`
        });

        suggestions.push({
            label: `helpers:convertCurrency("from", "to", value)`,
            kind: kind,
            insertText: `helpers:convertCurrency("from", "to", value)`,
            insertTextRules: snippet,
            documentation: `Convert currency from one to another using exchange rate`,
        });

        const simpleFields = ['title', 'destinationAmount', 'sourceAmount', 'sourceCurrency', 'destinationCurrency', 'sourceAccountID', 'destinationAccountID', 'notes', 'transactionType', 'referenceNumber', 'internalReferenceNumber'];

        const intFields = new Set(['sourceAmount', 'destinationAmount', 'sourceAccountID', 'destinationAccountID', 'transactionType']);

        for (let field of simpleFields) {
            suggestions.push({
                label: `tx:${field}()`,
                kind: kind,
                insertText: `tx:${field}`,
                documentation: `Get value of ${field} field`
            });

            const isInt = intFields.has(field);
            const insertText = isInt ? `tx:${field}(\${1:value})` : `tx:${field}("\${1:value}")`;

            const label = isInt ? `tx:${field}(value)` : `tx:${field}("value")`;

            suggestions.push({
                label,
                kind: kind,
                insertText,
                insertTextRules: snippet,
                documentation: `Set value of ${field} field`
            });
        }

        for (let field of ['getDestinationAmountWithDecimalPlaces', 'getSourceAmountWithDecimalPlaces']) {
            suggestions.push({
                label: `tx:${field}("value")`,
                kind: kind,
                insertText: `tx:${field}(\${1:value})`,
                insertTextRules: snippet,
                documentation: `Get value of ${field} field with decimal places`
            });
        }

        suggestions.push({
            label: `tx:addTag(<tagID>)`,
            kind: kind,
            insertText: `tx:addTag(\${1:value})`,
            insertTextRules: snippet,
            documentation: `Add tag to transaction`
        });

        suggestions.push({
            label: `tx:removeTag(<tagID>)`,
            kind: kind,
            insertText: `tx:removeTag(\${1:value})`,
            insertTextRules: snippet,
            documentation: `Remove tag from transaction`
        });

        suggestions.push({
            label: `tx:getTags()`,
            kind: kind,
            insertText: `tx:getTags()`,
            insertTextRules: snippet,
            documentation: `Get list of tags for transaction (lua map)`
        });

        suggestions.push({
            label: `tx:removeAllTags()`,
            kind: kind,
            insertText: `tx:removeAllTags()`,
            insertTextRules: snippet,
            documentation: `Remove all tags from transaction`
        });

        return suggestions;
    }

    private monacoRegistered = false;

    onEditorInit($event: IStandaloneCodeEditor) {
        let monaco = (window as any).monaco;

        if (this.monacoRegistered)
            return;

        this.monacoRegistered = true;
        monaco.languages.registerCompletionItemProvider('lua', {
            triggerCharacters: [':', '.'],
            provideCompletionItems: () => {
                return {
                    suggestions: this.getSuggestions()
                };
            }
        });
    }
}
