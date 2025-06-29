import { Component, ElementRef, Inject, OnInit, ViewChild } from '@angular/core';
import { Table, TableModule } from 'primeng/table';
import { FormsModule } from '@angular/forms';
import { InputText } from 'primeng/inputtext';
import { ToastModule } from 'primeng/toast';
import { InputIcon } from 'primeng/inputicon';
import { IconField } from 'primeng/iconfield';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { Transport, createClient } from '@connectrpc/connect';
import { AccountsService, ListAccountsResponse_AccountItem } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { ErrorHelper } from '../../helpers/error.helper';
import { FilterMetadata, MessageService } from 'primeng/api';
import { CommonModule, DatePipe } from '@angular/common';
import { TimestampHelper } from '../../helpers/timestamp.helper';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { Button } from 'primeng/button';
import { EnumService, AccountTypeEnum } from '../../services/enum.service';
import { MultiSelectModule } from 'primeng/multiselect';
import { SelectModule } from 'primeng/select';
import { OverlayModule } from 'primeng/overlay';
import { ListTagsResponse_TagItem, TagsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/tags/v1/tags_pb';
import { RulesService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/rules/v1/rules_pb';
import { Rule } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/rule_pb';

class RuleGroup {
    title: string = 'Default';
    rules: Rule[] = [];
}

@Component({
    selector: 'app-rule-list',
    templateUrl: 'rule-list.component.html',
    imports: [OverlayModule, FormsModule, InputText, ToastModule, TableModule, InputIcon, IconField, Button, MultiSelectModule, SelectModule, CommonModule, RouterLink],
    styles: `
        :host ::ng-deep .tagListingTable .p-datatable-header {
            border-width: 0 !important;
        }
    `
})
export class RuleListComponent implements OnInit {
    statuses: any[] = [];

    loading: boolean = false;

    public rules: Rule[] = [];
    public ruleGroups: RuleGroup[] = [];

    private ruleService;

    public filters: { [s: string]: FilterMetadata } = {};

    @ViewChild('filter') filter!: ElementRef;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        public router: Router
    ) {
        this.ruleService = createClient(RulesService, this.transport);
    }

    getAccountUrl(rule: Rule): string {
        return this.router.createUrlTree(['/', 'rules', rule!.id.toString()]).toString();
    }

    async ngOnInit() {
        this.loading = true;

        try {
            let resp = await this.ruleService.listRules({});
            this.rules = resp.rules || [];

            let ruleMap: { [key: string]: Rule[] } = {};
            for (const rule of this.rules) {
                const groupName = rule.groupName || 'Default';
                if (!ruleMap[groupName]) {
                    ruleMap[groupName] = [];
                }

                ruleMap[groupName].push(rule);
            }

            this.ruleGroups = Object.keys(ruleMap).map((groupName) => {
                return {
                    title: groupName,
                    rules: ruleMap[groupName]
                };
            });

            console.log(this.ruleGroups)
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

    protected readonly TimestampHelper = TimestampHelper;
}
