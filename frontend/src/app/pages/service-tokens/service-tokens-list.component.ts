import { Component, ElementRef, Inject, OnInit, ViewChild } from '@angular/core';
import { Table, TableModule } from 'primeng/table';
import { FormsModule } from '@angular/forms';
import { InputText } from 'primeng/inputtext';
import { ToastModule } from 'primeng/toast';
import { InputIcon } from 'primeng/inputicon';
import { IconField } from 'primeng/iconfield';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { Transport, createClient } from '@connectrpc/connect';
import { ErrorHelper } from '../../helpers/error.helper';
import { ConfirmationService, MessageService } from 'primeng/api';
import { CommonModule } from '@angular/common';
import { TimestampHelper } from '../../helpers/timestamp.helper';
import { Router } from '@angular/router';
import { Button } from 'primeng/button';
import { TagModule } from 'primeng/tag';
import { ConfirmDialogModule } from 'primeng/confirmdialog';
import { DialogModule } from 'primeng/dialog';
import { ServiceToken } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/servicetoken_pb';
import { ConfigurationService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/configuration/v1/configuration_pb';
import { ServiceTokensCreateComponent } from './service-tokens-create.component';
import { TooltipModule } from 'primeng/tooltip';
import { Message } from 'primeng/message';

@Component({
    selector: 'app-service-tokens-list',
    templateUrl: 'service-tokens-list.component.html',
    imports: [
        FormsModule,
        InputText,
        ToastModule,
        TableModule,
        InputIcon,
        IconField,
        Button,
        CommonModule,
        TagModule,
        ConfirmDialogModule,
        DialogModule,
        ServiceTokensCreateComponent,
        TooltipModule,
        Message
    ],
    providers: [ConfirmationService],
    styles: `
        :host ::ng-deep .serviceTokensTable .p-datatable-header {
            border-width: 0 !important;
        }
    `
})
export class ServiceTokensListComponent implements OnInit {
    loading: boolean = false;
    tokens: ServiceToken[] = [];
    private configService;
    showCreateDialog: boolean = false;
    createdToken: string = '';
    showTokenDialog: boolean = false;

    @ViewChild('filter') filter!: ElementRef;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        private confirmationService: ConfirmationService,
        public router: Router
    ) {
        this.configService = createClient(ConfigurationService, this.transport);
    }

    async ngOnInit() {
        await this.loadTokens();
    }

    async loadTokens() {
        this.loading = true;

        try {
            let resp = await this.configService.getServiceTokens({});
            this.tokens = resp.serviceTokens || [];
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        } finally {
            this.loading = false;
        }
    }

    onGlobalFilter(table: Table, event: Event) {
        table.filterGlobal((event.target as HTMLInputElement).value, 'contains');
    }

    clear(table: Table) {
        table.clear();
        this.filter.nativeElement.value = '';
    }

    isExpired(token: ServiceToken): boolean {
        if (!token.expiresAt) return false;
        const expiresAt = TimestampHelper.timestampToDate(token.expiresAt);
        return expiresAt < new Date();
    }

    isRevoked(token: ServiceToken): boolean {
        return !!token.deletedAt;
    }

    getStatus(token: ServiceToken): string {
        if (this.isRevoked(token)) return 'Revoked';
        if (this.isExpired(token)) return 'Expired';
        return 'Active';
    }

    getStatusSeverity(token: ServiceToken): 'success' | 'info' | 'warn' | 'danger' | 'secondary' | 'contrast' {
        if (this.isRevoked(token)) return 'danger';
        if (this.isExpired(token)) return 'warn';
        return 'success';
    }

    confirmRevoke(token: ServiceToken) {
        this.confirmationService.confirm({
            message: `Are you sure you want to revoke the token "${token.name}"? This action cannot be undone.`,
            header: 'Confirm Revoke',
            icon: 'pi pi-exclamation-triangle',
            accept: async () => {
                await this.revokeToken(token);
            }
        });
    }

    async revokeToken(token: ServiceToken) {
        try {
            await this.configService.revokeServiceToken({ id: token.id });
            this.messageService.add({ severity: 'success', detail: 'Token revoked successfully' });
            await this.loadTokens();
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    openCreateDialog() {
        this.showCreateDialog = true;
    }

    onTokenCreated(token: string) {
        this.showCreateDialog = false;
        this.createdToken = token;
        this.showTokenDialog = true;
        this.loadTokens();
    }

    copyToken() {
        navigator.clipboard.writeText(this.createdToken);
        this.messageService.add({ severity: 'success', detail: 'Token copied to clipboard' });
    }

    closeTokenDialog() {
        this.showTokenDialog = false;
        this.createdToken = '';
    }

    protected readonly TimestampHelper = TimestampHelper;
}
