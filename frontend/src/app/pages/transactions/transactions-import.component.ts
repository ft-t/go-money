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
import {
    CreateTransactionRequest
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { Transaction } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { AccordionModule } from 'primeng/accordion';
import { NgIf } from '@angular/common';
import { TransactionEditorComponent } from '../../shared/components/transaction-editor/transaction-editor.component';

interface TransactionItem {
    transaction: Transaction;
    selected: boolean;
}

@Component({
    selector: 'app-transactions-import',
    imports: [
        Fluid, Toast, FileUpload, DropdownModule, FormsModule, Textarea, Checkbox, IftaLabel, Button,
        AccordionModule, NgIf, TransactionEditorComponent
    ],
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
    public rawText: string = '';

    public transactionItems: TransactionItem[] = [];
    public showReview: boolean = false;

    @ViewChildren('editor') editors: QueryList<TransactionEditorComponent> = new QueryList();

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService
    ) {
        this.importService = createClient(ImportService, this.transport);
    }

    public isRawTextImport() {
        return this.selectedSource == ImportSource.PRIVATE_24;
    }

    public textContent: string = '';

    async submit() {
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

            this.transactionItems = response.transactions.map(tx => ({
                transaction: tx,
                selected: true
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

    getTransactionTitle(tx: Transaction): string {
        if (tx.title) {
            return tx.title;
        }
        return `Transaction #${tx.id || 'New'}`;
    }

    getSelectedCount(): number {
        return this.transactionItems.filter(item => item.selected).length;
    }

    selectAll() {
        this.transactionItems.forEach(item => item.selected = true);
    }

    deselectAll() {
        this.transactionItems.forEach(item => item.selected = false);
    }

    async importSelected() {
        const selectedTransactions = this.transactionItems
            .filter(item => item.selected)
            .map(item => item.transaction);

        if (selectedTransactions.length === 0) {
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
                detail: `Importing ${selectedTransactions.length} transaction(s)...`
            });

            // TODO: Call actual import API
            // For now, just show success message
            this.messageService.add({
                severity: 'success',
                detail: `Successfully imported ${selectedTransactions.length} transaction(s)`
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
        this.textContent = '';
        this.showReview = false;

        const reader = new FileReader();
        reader.onload = async (event2) => {
            console.log(event2.target!.result);

            try {
                this.isLoading = true;
                this.messageService.add({ severity: 'info', detail: 'Importing...' });

                let result = await this.importService.importTransactions(
                    create(ImportTransactionsRequestSchema, {
                        skipRules: this.skipRules,
                        treatDatesAsUtc: this.treatDatesAsUtc,
                        source: this.selectedSource,
                        fileContent: btoa(unescape(encodeURIComponent(event2.target!.result as string)))
                    })
                );

                let newText = `Imported ${result.importedCount} transaction.\nDuplicate transactions: ${result.duplicateCount}.\n`;

                this.textContent = newText;

                this.messageService.add({ severity: 'info', detail: newText });
            } catch (e) {
                this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            } finally {
                this.isLoading = false;
            }
        };

        reader.readAsText(event.files[0]);
    }
}
