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
import { NgIf } from '@angular/common';
import { TransactionEditorComponent } from '../../shared/components/transaction-editor/transaction-editor.component';
import { Transaction } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { TransactionsService, CreateTransactionsBulkRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';

interface TransactionItem {
    transaction: Transaction;
    selected: boolean;
    duplicateTxID?: bigint;
}

@Component({
    selector: 'app-transactions-import',
    imports: [Fluid, Toast, FileUpload, DropdownModule, FormsModule, Textarea, Checkbox, IftaLabel, Button, AccordionModule, NgIf, TransactionEditorComponent],
    templateUrl: './transactions-import.component.html'
})
export class TransactionsImportComponent {
    public selectedSource: ImportSource = ImportSource.PRIVATE_24;
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

    public textContent: string = '';

    async submit() {
        if (this.stagingMode) {
            await this.parseForReview();
        } else {
            await this.directImport();
        }
    }

    async parseForReview() {
        this.transactionItems = [];
        this.textContent = '';
        this.showReview = false;

        try {
            this.isLoading = true;
            this.messageService.add({ severity: 'info', detail: 'Parsing...' });

            const response = await this.importService.parseTransactions(
                create(ParseTransactionsRequestSchema, {
                    content: [this.rawText],
                    source: this.selectedSource,
                    treatDatesAsUtc: this.treatDatesAsUtc
                })
            );

            this.transactionItems = response.transactions.map((tx) => ({
                transaction: tx.transaction!,
                selected: tx.duplicateTransactionId === undefined,
                duplicateTxID: tx.duplicateTransactionId
            }));

            if (this.transactionItems.length > 0) {
                this.showReview = true;
                this.messageService.add({
                    severity: 'success',
                    detail: `Parsed ${this.transactionItems.length} transaction(s)`
                });
            } else {
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

    async directImport() {
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
                    content: [this.rawText]
                })
            );

            const logText = `Imported ${result.importedCount} transaction(s).\nDuplicate transactions: ${result.duplicateCount}.\n`;
            this.textContent = logText;
            this.messageService.add({ severity: 'success', detail: logText });
            this.rawText = '';
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
        return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
    }

    formatAmount(amount: string, currency: string): string {
        return `${amount} ${currency}`;
    }

    getSelectedCount(): number {
        return this.transactionItems.filter((item) => item.selected && item.duplicateTxID === undefined).length;
    }

    getNonDuplicateCount(): number {
        return this.transactionItems.filter((item) => item.duplicateTxID === undefined).length;
    }

    selectAll() {
        this.transactionItems.forEach((item) => {
            if (item.duplicateTxID === undefined) {
                item.selected = true;
            }
        });
    }

    deselectAll() {
        this.transactionItems.forEach((item) => {
            if (item.duplicateTxID === undefined) {
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

    getFilteredTransactionItems(): TransactionItem[] {
        if (this.hideDuplicates) {
            return this.transactionItems.filter((item) => item.duplicateTxID === undefined);
        }
        return this.transactionItems;
    }

    async importSelected() {
        const selectedIndices: number[] = [];

        // Get indices of selected items that are not duplicates
        this.transactionItems.forEach((item, index) => {
            if (item.selected && item.duplicateTxID === undefined) {
                selectedIndices.push(index);
            }
        });

        if (selectedIndices.length === 0) {
            this.messageService.add({
                severity: 'warn',
                detail: 'No transactions selected'
            });
            return;
        }

        try {
            this.isLoading = true;
            this.messageService.add({
                severity: 'info',
                detail: `Importing ${selectedIndices.length} transaction(s)...`
            });

            // Build transaction requests from editors
            const transactionRequests = selectedIndices.map(index => {
                const editor = this.editors.toArray()[index];
                return editor.buildTransactionRequest();
            });

            // Call bulk create API
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
                await this.parseForReviewWithContent(content);
            } else {
                await this.directImportWithContent(content);
            }
        };

        reader.readAsText(event.files[0]);
    }

    async parseForReviewWithContent(content: string) {
        this.transactionItems = [];
        this.textContent = '';
        this.showReview = false;

        try {
            this.isLoading = true;
            this.messageService.add({ severity: 'info', detail: 'Parsing...' });

            const response = await this.importService.parseTransactions(
                create(ParseTransactionsRequestSchema, {
                    content: [content],
                    source: this.selectedSource,
                    treatDatesAsUtc: this.treatDatesAsUtc
                })
            );

            this.transactionItems = response.transactions.map((tx) => ({
                transaction: tx.transaction!,
                selected: tx.duplicateTransactionId === undefined,
                duplicateTxID: tx.duplicateTransactionId
            }));

            if (this.transactionItems.length > 0) {
                this.showReview = true;
                this.messageService.add({
                    severity: 'success',
                    detail: `Parsed ${this.transactionItems.length} transaction(s)`
                });
            } else {
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

    async directImportWithContent(content: string) {
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
                    content: [content]
                })
            );

            const logText = `Imported ${result.importedCount} transaction(s).\nDuplicate transactions: ${result.duplicateCount}.\n`;
            this.textContent = logText;
            this.messageService.add({ severity: 'success', detail: logText });
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        } finally {
            this.isLoading = false;
        }
    }
}
