import { Component, Inject, Input } from '@angular/core';
import { InputNumberModule } from 'primeng/inputnumber';
import { TextareaModule } from 'primeng/textarea';
import { Fluid } from 'primeng/fluid';
import { InputText } from 'primeng/inputtext';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { NgIf } from '@angular/common';
import { Button } from 'primeng/button';
import { Toast } from 'primeng/toast';
import { ColorPickerModule } from 'primeng/colorpicker';
import { Checkbox } from 'primeng/checkbox';
import { DiffEditorComponent, DiffEditorModel, EditorComponent } from 'ngx-monaco-editor-v2';
import { editor } from 'monaco-editor';
import IStandaloneCodeEditor = editor.IStandaloneCodeEditor;
import { ErrorHelper } from '../../../helpers/error.helper';
import { DryRunRuleRequestSchema, RulesService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/rules/v1/rules_pb';
import { createClient, Transport } from '@connectrpc/connect';
import { TransactionsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { TRANSPORT_TOKEN } from '../../../consts/transport';
import { MessageService } from 'primeng/api';
import { create } from '@bufbuild/protobuf';
import { RuleSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/rule_pb';

@Component({
    selector: 'app-script-editor',
    imports: [InputNumberModule, TextareaModule, InputText, ReactiveFormsModule, FormsModule, Button, ColorPickerModule, EditorComponent, DiffEditorComponent, NgIf],
    templateUrl: './script-editor.component.html'
})
export class ScriptEditorComponent {
    private monacoRegistered = false;
    editorOptions = { theme: 'vs-dark', language: 'lua' };
    diffOptions = {
        theme: 'vs-dark'
    };
    originalModel: DiffEditorModel = {
        code: '',
        language: 'text/json'
    };

    originalTextModel: editor.ITextModel | null = null;
    modifiedTextModel: editor.ITextModel | null = null;

    public modifiedModel: DiffEditorModel = {
        code: '',
        language: 'text/json'
    };

    public helpContent = this.generateSimpleApiHelp(this.getSuggestions());
    public dryRunTransactionId: number = 0;
    private transactionService;
    private rulesService;

    @Input() script: string = '';
    @Input() useEmptyTx: boolean = false;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService
    ) {
        this.rulesService = createClient(RulesService, this.transport);
        this.transactionService = createClient(TransactionsService, this.transport);
    }

    onEditorInit($event: IStandaloneCodeEditor) {
        let monaco = (window as any).monaco;

        if (this.monacoRegistered) return;

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

        if (this.useEmptyTx) {
            this.dryRunTransactionId = 0;
            return;
        }

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
            documentation: `Convert currency from one to another using exchange rate`
        });

        const simpleFields = ['title', 'categoryID', 'destinationAmount', 'sourceAmount', 'sourceCurrency', 'destinationCurrency', 'sourceAccountID', 'destinationAccountID', 'notes', 'transactionType', 'referenceNumber'];

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

        suggestions.push({
            label: `tx:transactionDateTimeSetTime(12,5)`,
            kind: kind,
            insertText: `tx:transactionDateTimeSetTime(\${1:value},\${2:value})`,
            insertTextRules: snippet,
            documentation: `Set transaction time to 12:05`
        });

        suggestions.push({
            label: `tx:transactionDateTimeAddDate(2,3,4)`,
            kind: kind,
            insertText: `tx:transactionDateTimeAddDate(\${1:value},\${2:value},\${2:value})`,
            insertTextRules: snippet,
            documentation: `Adds 2 years, 3 months and 4 days to transaction date`
        });

        suggestions.push({
            label: `tx:getInternalReferenceNumbers()`,
            kind: kind,
            insertText: `tx:getInternalReferenceNumbers()`,
            documentation: `Get all internal reference numbers as a Lua table`
        });

        suggestions.push({
            label: `tx:addInternalReferenceNumber("value")`,
            kind: kind,
            insertText: `tx:addInternalReferenceNumber("\${1:value}")`,
            insertTextRules: snippet,
            documentation: `Add a new internal reference number`
        });

        suggestions.push({
            label: `tx:setInternalReferenceNumbers(table)`,
            kind: kind,
            insertText: `tx:setInternalReferenceNumbers({"\${1:value1}", "\${2:value2}"})`,
            insertTextRules: snippet,
            documentation: `Replace all internal reference numbers with given table`
        });

        suggestions.push({
            label: `tx:removeInternalReferenceNumber("value")`,
            kind: kind,
            insertText: `tx:removeInternalReferenceNumber("\${1:value}")`,
            insertTextRules: snippet,
            documentation: `Remove a specific internal reference number`
        });

        return suggestions;
    }

    async dryRun() {
        await this.ensureTxSet();

        try {
            let response = await this.rulesService.dryRunRule(
                create(DryRunRuleRequestSchema, {
                    rule: create(RuleSchema, {
                        script: this.script
                    }),
                    transactionId: BigInt(+this.dryRunTransactionId)
                })
            );

            this.originalTextModel!.setValue(JSON.stringify(response.before, null, 2));

            this.modifiedTextModel!.setValue(JSON.stringify(response.after, null, 2));
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            throw e;
        }

        return this.script;
    }

    onDiffEditorInit(diffEditor: any) {
        const model = diffEditor.getModel();
        this.originalTextModel = model.original;
        this.modifiedTextModel = model.modified;
    }
}
