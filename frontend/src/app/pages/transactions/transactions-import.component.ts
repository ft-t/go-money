import { Component, Inject, QueryList, ViewChildren } from '@angular/core';
import { Fluid } from 'primeng/fluid';
import { Toast } from 'primeng/toast';
import { FileUpload } from 'primeng/fileupload';
import { ImportService, ImportSource, ImportTransactionsRequestSchema, ParseTransactionsRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/import/v1/import_pb';
import { EnumService } from '../../services/enum.service';
import { DropdownModule } from 'primeng/dropdown';
import { FormsModule } from '@angular/forms';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { create } from '@bufbuild/protobuf';
import { Textarea } from 'primeng/textarea';
import { ErrorHelper } from '../../helpers/error.helper';
import { MessageService } from 'primeng/api';
import { Checkbox } from 'primeng/checkbox';
import { IftaLabel } from 'primeng/iftalabel';
import { Button } from 'primeng/button';
import { AccordionModule } from 'primeng/accordion';
import { NgClass, NgIf } from '@angular/common';
import { TransactionEditorComponent } from '../../shared/components/transaction-editor/transaction-editor.component';
import { Message } from 'primeng/message';
import { Transaction, TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { TransactionsService, CreateTransactionsBulkRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';

interface TransactionItem {
    transaction: Transaction;
    selected: boolean;
    duplicateTxID?: bigint;
    hasError: boolean;
    hasValidationError?: boolean;
}

@Component({
    selector: 'app-transactions-import',
    imports: [Fluid, Toast, FileUpload, DropdownModule, FormsModule, Textarea, Checkbox, IftaLabel, Button, AccordionModule, NgIf, NgClass, TransactionEditorComponent, Message],
    templateUrl: './transactions-import.component.html',
    styles: [`
        :host ::ng-deep .validation-error .p-accordiontab-header {
            border-left: 4px solid #ef4444 !important;
            background-color: #fef2f2 !important;
        }

        :host ::ng-deep .validation-error .p-accordiontab-header:hover {
            background-color: #fee2e2 !important;
        }
    `]
})
export class TransactionsImportComponent {
    public selectedSource: ImportSource = ImportSource.FIREFLY;
    public sources = EnumService.getImportTypes();
    public skipRules: boolean = false;
    public treatDatesAsUtc: boolean = false;
    public isLoading: boolean = false;
    public stagingMode: boolean = true;

    public importService;
    public transactionService;
    public rawText: string = '';

    public transactionItems: TransactionItem[] = [];
    public showReview: boolean = false;
    public hideDuplicates: boolean = false;
    public showErrorsOnly: boolean = false;
    public activeIndices: number[] = [];

    @ViewChildren('editor') editors: QueryList<TransactionEditorComponent> = new QueryList();

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService
    ) {
        this.importService = createClient(ImportService, this.transport);
        this.transactionService = createClient(TransactionsService, this.transport);
    }

    public isRawTextImport() {
        return this.selectedSource == ImportSource.PRIVATE_24;
    }

    public getImportNotes(): string | null {
        switch (this.selectedSource) {
            case ImportSource.MONOBANK:
                return 'For Monobank imports, account names should be in format: Monobank_<CURRENCY> (e.g., Monobank_UAH, Monobank_USD)';
            default:
                return null;
        }
    }

    public textContent: string = '';

    async submit() {
        if (this.stagingMode) {
            await this.parseForReview();
        } else {
            await this.directImport();
        }
    }

    async parseForReview(content?: string) {
        this.transactionItems = [];
        this.textContent = '';
        this.showReview = false;

        try {
            this.isLoading = true;
            this.messageService.add({ severity: 'info', detail: 'Parsing...' });

            const response = await this.importService.parseTransactions(
                create(ParseTransactionsRequestSchema, {
                    content: [content || this.rawText],
                    source: this.selectedSource,
                    treatDatesAsUtc: this.treatDatesAsUtc
                })
            );

            this.transactionItems = response.transactions.map((tx) => ({
                transaction: tx.transaction!,
                selected: tx.duplicateTransactionId === undefined && tx.transaction!.type !== TransactionType.UNSPECIFIED,
                duplicateTxID: tx.duplicateTransactionId,
                hasError: tx.transaction!.type === TransactionType.UNSPECIFIED
            }));

            this.transactionItems = this.transactionItems.sort((a, b) => {
                if (a.hasError && !b.hasError) return -1;
                if (!a.hasError && b.hasError) return 1;

                const dateA = a.transaction.transactionDate?.seconds || 0n;
                const dateB = b.transaction.transactionDate?.seconds || 0n;
                return dateA > dateB ? -1 : dateA < dateB ? 1 : 0;
            });

            if (this.transactionItems.length > 0) {
                this.showReview = true;
                const duplicateCount = this.getDuplicatesCount();
                const errorCount = this.getErrorCount();
                const logText = `Parsed ${this.transactionItems.length} transaction(s).\nDuplicate transactions: ${duplicateCount}.\nParsing errors: ${errorCount}.`;
                this.textContent = logText;
                this.messageService.add({
                    severity: 'success',
                    detail: `Parsed ${this.transactionItems.length} transaction(s)`
                });
            } else {
                this.textContent = 'No transactions found.';
                this.messageService.add({
                    severity: 'info',
                    detail: 'No transactions found'
                });
            }
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        } finally {
            this.isLoading = false;
        }
    }

    async directImport(content?: string) {
        this.textContent = '';
        this.showReview = false;

        try {
            this.isLoading = true;
            this.messageService.add({ severity: 'info', detail: 'Importing...' });

            const result = await this.importService.importTransactions(
                create(ImportTransactionsRequestSchema, {
                    skipRules: this.skipRules,
                    treatDatesAsUtc: this.treatDatesAsUtc,
                    source: this.selectedSource,
                    content: [content || this.rawText]
                })
            );

            const logText = `Imported ${result.importedCount} transaction(s).\nDuplicate transactions: ${result.duplicateCount}.\n`;
            this.textContent = logText;
            this.messageService.add({ severity: 'success', detail: logText });
            if (!content) {
                this.rawText = '';
            }
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        } finally {
            this.isLoading = false;
        }
    }

    getTransactionTitle(tx: Transaction): string {
        if (tx.title) {
            return tx.title;
        }
        return `Transaction #${tx.id || 'New'}`;
    }

    getTransactionTypeLabel(type: number): string {
        const transactionType = EnumService.getAllTransactionTypes().find(t => t.value === type);
        return transactionType?.name || 'Unknown';
    }

    formatDate(timestamp?: any): string {
        if (!timestamp) return '';
        const date = new Date(Number(timestamp.seconds) * 1000);
        return date.toISOString().slice(0, 16).replace('T', ' ');
    }

    formatAmount(amount: string, currency: string): string {
        return `${amount} ${currency}`;
    }

    getSelectedCount(): number {
        return this.transactionItems.filter((item) => item.selected && item.duplicateTxID === undefined && !item.hasError).length;
    }

    getNonDuplicateCount(): number {
        return this.transactionItems.filter((item) => item.duplicateTxID === undefined && !item.hasError).length;
    }

    getErrorCount(): number {
        return this.transactionItems.filter((item) => item.hasError).length;
    }

    selectAll() {
        this.transactionItems.forEach((item) => {
            if (item.duplicateTxID === undefined && !item.hasError) {
                item.selected = true;
            }
        });
    }

    deselectAll() {
        this.transactionItems.forEach((item) => {
            if (item.duplicateTxID === undefined && !item.hasError) {
                item.selected = false;
            }
        });
    }

    getDuplicatesCount(): number {
        return this.transactionItems.filter((item) => item.duplicateTxID !== undefined).length;
    }

    toggleHideDuplicates() {
        this.hideDuplicates = !this.hideDuplicates;
    }

    toggleShowErrorsOnly() {
        this.showErrorsOnly = !this.showErrorsOnly;
        if (!this.showErrorsOnly) {
            this.hideDuplicates = false;
        }
    }

    getFilteredTransactionItems(): TransactionItem[] {
        let filtered = this.transactionItems;

        if (this.showErrorsOnly) {
            return filtered.filter((item) => item.hasError);
        }

        if (this.hideDuplicates) {
            return filtered.filter((item) => item.duplicateTxID === undefined);
        }

        return filtered;
    }

    async importSelected() {
        if (this.transactionItems.filter(item => item.selected && !item.hasError && item.duplicateTxID === undefined).length === 0) {
            this.messageService.add({
                severity: 'warn',
                detail: 'No transactions selected'
            });
            return;
        }

        this.transactionItems.forEach(item => item.hasValidationError = false);

        let hasValidationErrors = false;
        const editorArray = this.editors.toArray();
        const filteredItems = this.getFilteredTransactionItems();
        let editorIndex = 0;

        filteredItems.forEach((item, filteredIndex) => {
            if (item.hasError) {
                return;
            }

            const editor = editorArray[editorIndex];

            if (item.selected && item.duplicateTxID === undefined) {
                if (editor && !editor.isValid()) {
                    hasValidationErrors = true;
                    item.hasValidationError = true;
                }
            }

            editorIndex++;
        });

        if (hasValidationErrors) {
            this.activeIndices = [];
            filteredItems.forEach((item, index) => {
                if (item.hasValidationError) {
                    this.activeIndices.push(index);
                }
            });

            this.messageService.add({
                severity: 'error',
                detail: 'Please fix validation errors before importing'
            });
            return;
        }

        try {
            this.isLoading = true;
            this.messageService.add({
                severity: 'info',
                detail: `Importing ${editorArray.length} transaction(s)...`
            });

            const transactionRequests = editorArray
                .filter((editor, index) => {
                    let editorIdx = 0;
                    for (let i = 0; i < filteredItems.length; i++) {
                        const item = filteredItems[i];
                        if (item.hasError) continue;
                        if (editorIdx === index) {
                            return item.selected && item.duplicateTxID === undefined;
                        }
                        editorIdx++;
                    }
                    return false;
                })
                .map(editor => editor.buildTransactionRequest());

            const response = await this.transactionService.createTransactionsBulk(
                create(CreateTransactionsBulkRequestSchema, {
                    transactions: transactionRequests
                })
            );

            this.messageService.add({
                severity: 'success',
                detail: `Successfully imported ${response.transactions.length} transaction(s)`
            });

            this.showReview = false;
            this.transactionItems = [];
            this.rawText = '';
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        } finally {
            this.isLoading = false;
        }
    }

    async onBasicUploadAuto(event: any) {
        const reader = new FileReader();
        reader.onload = async (event2) => {
            const content = btoa(unescape(encodeURIComponent(event2.target!.result as string)));

            if (this.stagingMode) {
                await this.parseForReview(content);
            } else {
                await this.directImport(content);
            }
        };

        reader.readAsText(event.files[0]);
    }


}
