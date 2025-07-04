import { Component, Inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterModule } from '@angular/router';
import { MenuItem, MenuItemCommandEvent, MessageService } from 'primeng/api';
import { AppMenuitem } from './app.menuitem';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { ConfigurationService, GetConfigurationResponse, GetConfigurationResponseSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/configuration/v1/configuration_pb';
import { ErrorHelper } from '../../helpers/error.helper';
import { create } from '@bufbuild/protobuf';

@Component({
    selector: 'app-menu',
    standalone: true,
    imports: [CommonModule, AppMenuitem, RouterModule],
    template: `
        <ul class="layout-menu">
            <ng-container *ngFor="let item of model; let i = index">
                <li app-menuitem *ngIf="!item.separator" [item]="item" [index]="i" [root]="true"></li>
                <li *ngIf="item.separator" class="menu-separator"></li>
            </ng-container>
        </ul>
    `
})
export class AppMenu implements OnInit {
    model: MenuItem[] = [];
    public configService;
    public config: GetConfigurationResponse = create(GetConfigurationResponseSchema, {});

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        private router: Router
    ) {
        this.configService = createClient(ConfigurationService, this.transport);
    }

    async ngOnInit() {
        try {
            let resp = await this.configService.getConfiguration({});

            this.config = resp;
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }

        this.model = [
            {
                label: 'Home',
                items: [
                    {
                        label: 'Dashboard',
                        icon: 'pi pi-fw pi-home',
                        command: (event: MenuItemCommandEvent) => {
                            if (event.originalEvent) {
                                event.originalEvent.preventDefault();
                            }

                            if(this.config.grafanaUrl) {
                                window.open(this.config.grafanaUrl, "_blank");
                                return;
                            }

                            this.router.navigateByUrl('/dashboard')
                        },
                    }
                ]
            },
            {
                label: 'Accounts',
                items: [
                    {
                        label: 'Assets',
                        icon: 'pi pi-fw pi-wallet',
                        routerLink: ['/accounts']
                    },
                    {
                        label: 'Liabilities',
                        icon: 'pi pi-fw pi-credit-card',
                        routerLink: ['/accounts/liabilities']
                    }
                ]
            },
            {
                label: 'Transactions',
                items: [
                    {
                        label: 'Deposits',
                        icon: 'pi pi-fw pi-plus',
                        routerLink: ['/transactions/deposits']
                    },
                    {
                        label: 'Withdrawals',
                        icon: 'pi pi-fw pi-minus',
                        routerLink: ['/transactions/withdrawals']
                    },
                    {
                        label: 'Transfers',
                        icon: 'pi pi-fw pi-send',
                        routerLink: ['/transactions/transfers']
                    },
                    {
                        label: 'All',
                        icon: 'pi pi-fw pi-asterisk',
                        routerLink: ['/transactions']
                    }
                ]
            },
            {
                label: 'Bulk',
                items: [
                    {
                        label: 'Transactions Import',
                        icon: 'pi pi-fw pi-file-import',
                        routerLink: ['/transactions/import']
                    },
                    {
                        label: 'Accounts Import',
                        icon: 'pi pi-fw pi-globe',
                        routerLink: ['/accounts/import']
                    },
                    {
                        label: 'Tags Import',
                        icon: 'pi pi-fw pi-hashtag',
                        routerLink: ['/tags/import']
                    }
                ]
            },
            {
                label: 'Automation',
                items: [
                    {
                        label: 'Rules',
                        icon: 'pi pi-fw pi-file-import',
                        routerLink: ['/rules']
                    }
                ]
            },
            {
                label: 'Admin Section',
                items: [
                    {
                        label: 'Tags',
                        icon: 'pi pi-fw pi-tag',
                        routerLink: ['/tags']
                    }
                    // {
                    //     label: 'Currencies',
                    //     icon: 'pi pi-fw pi-table',
                    //     routerLink: ['/accounts']
                    // },
                    // {
                    //     label: 'Users',
                    //     icon: 'pi pi-fw pi-table',
                    //     routerLink: ['/accounts']
                    // },
                    // {
                    //     label: 'Debug',
                    //     icon: 'pi pi-fw pi-table',
                    //     routerLink: ['/accounts']
                    // }
                ]
            }
        ];

        console.log(this.model)
    }
}
