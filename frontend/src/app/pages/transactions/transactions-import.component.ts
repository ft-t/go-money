import { Component, Inject } from '@angular/core';
import { Fluid } from 'primeng/fluid';
import { Toast } from 'primeng/toast';
import { FileUpload } from 'primeng/fileupload';
import { ImportService, ImportSource, ImportTransactionsRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/import/v1/import_pb';
import { EnumService } from '../../services/enum.service';
import { DropdownModule } from 'primeng/dropdown';
import { FormsModule } from '@angular/forms';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { TransactionSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { TransactionsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/v1/transactions_pb';
import { create } from '@bufbuild/protobuf';
import { Textarea } from 'primeng/textarea';
import { ErrorHelper } from '../../helpers/error.helper';
import { MessageService } from 'primeng/api';
import { Checkbox } from 'primeng/checkbox';

@Component({
    selector: 'app-transactions-import',
    imports: [Fluid, Toast, FileUpload, DropdownModule, FormsModule, Textarea, Checkbox],
    templateUrl: './transactions-import.component.html'
})
export class TransactionsImportComponent {
    public selectedSource: ImportSource = ImportSource.FIREFLY;
    public sources = EnumService.getImportTypes();
    public skipRules: boolean = false;
    public isLoading: boolean = false;

    public importService;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService
    ) {
        this.importService = createClient(ImportService, this.transport);
    }

    public textContent: string = '';

    async onBasicUploadAuto(event: any) {
        this.textContent = '';

        const reader = new FileReader();
        reader.onload = async (event2) => {
            console.log(event2.target!.result);

            try {
                this.isLoading = true;
                this.messageService.add({ severity: 'info', detail: 'Importing...' });

                let result = await this.importService.importTransactions(
                    create(ImportTransactionsRequestSchema, {
                        skipRules: this.skipRules,
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
