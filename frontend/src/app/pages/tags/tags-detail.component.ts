import { Component, Inject, OnInit } from '@angular/core';
import { FilterWrapper, TransactionsTableComponent } from '../../shared/components/transactions-table/transactions-table.component';
import { Tag, TagSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/tag_pb';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { FilterMetadata, MessageService } from 'primeng/api';
import { ActivatedRoute, Router } from '@angular/router';
import { TagsService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/tags/v1/tags_pb';
import { create } from '@bufbuild/protobuf';
import { ErrorHelper } from '../../helpers/error.helper';
import { TransactionType } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';

@Component({
    selector: 'app-tags-detail',
    imports: [TransactionsTableComponent],
    templateUrl: './tags-detail.component.html'
})
export class TagsDetailComponent implements OnInit {
    public tag: Tag = create(TagSchema, {});
    private tagService;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        routeSnapshot: ActivatedRoute
    ) {
        this.tagService = createClient(TagsService, this.transport);

        this.tag.id = +routeSnapshot.snapshot.params['id'];
    }

    async ngOnInit() {
        try {
            let response = await this.tagService.listTags({ ids: [+this.tag.id] });
            if (response.tags && response.tags.length == 0) {
                this.messageService.add({ severity: 'error', detail: 'tag not found' });
                return;
            }

            this.tag = response.tags[0].tag ?? create(TagSchema, {});
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    getTableTitle() {
        return `Transactions for tag "${this.tag.name}"`;
    }

    getFilter(): FilterWrapper {
        return {
            filters: {
                'tags': {
                    matchMode: 'in',
                    value: [this.tag.id]
                }
            }
        };
    }
}
