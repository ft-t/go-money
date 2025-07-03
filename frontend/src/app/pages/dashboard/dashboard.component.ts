import { Component, Inject, OnInit } from '@angular/core';
import { Message } from 'primeng/message';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { create } from '@bufbuild/protobuf';
import { ErrorHelper } from '../../helpers/error.helper';
import { createClient, Transport } from '@connectrpc/connect';
import {
    ConfigurationService,
    GetConfigurationResponse, GetConfigurationResponseSchema
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/configuration/v1/configuration_pb';
import { MessageService } from 'primeng/api';
import { NgIf } from '@angular/common';
import { Button } from 'primeng/button';

@Component({
    selector: 'app-dashboard',
    imports: [Message, NgIf, Button],
    templateUrl: './dashboard.component.html'
})
export class DashboardComponent implements OnInit {
    public configService;
    public isLoading: boolean = true;
    public config: GetConfigurationResponse = create(GetConfigurationResponseSchema, {});

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService
    ) {
        this.configService = createClient(ConfigurationService, this.transport);
    }

    navigate() {
        window.open(this.config.grafanaUrl, "_blank");
    }

    async ngOnInit() {
        try {
            let resp = await this.configService.getConfiguration({});

            this.config = resp;
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        } finally {
            this.isLoading = false;
        }
    }
}
