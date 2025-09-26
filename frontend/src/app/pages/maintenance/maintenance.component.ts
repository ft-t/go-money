import { Component, Inject } from '@angular/core';
import { Fluid } from 'primeng/fluid';
import { ReactiveFormsModule } from '@angular/forms';
import { Button } from 'primeng/button';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { MaintenanceService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/maintenance/v1/maintenance_pb';
import { MessageService } from 'primeng/api';
import { ErrorHelper } from '../../helpers/error.helper';

@Component({
    selector: 'maintenance-component',
    imports: [Fluid, ReactiveFormsModule, Button],
    templateUrl: './maintenance.component.html'
})
export class MaintenanceComponent {
    private maintenanceService;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService
    ) {
        this.maintenanceService = createClient(MaintenanceService, this.transport);
    }

    async recalculateAll() {
        this.messageService.add({ severity: 'info', detail: 'Recalculation started. This may take a while...' });
        try{
            await this.maintenanceService.recalculateAll({});
            this.messageService.add({ severity: 'success', detail: 'Recalculation task has been completed' });
        }
        catch (e) {
            console.log(e);
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }
}
