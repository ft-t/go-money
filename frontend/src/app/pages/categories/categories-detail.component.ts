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
import { CategoriesService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/categories/v1/categories_pb';
import { Category, CategorySchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/category_pb';

@Component({
    selector: 'app-categories-detail',
    imports: [TransactionsTableComponent],
    templateUrl: './categories-detail.component.html'
})
export class CategoriesDetailComponent implements OnInit {
    public category: Category = create(CategorySchema, {});
    private categoriesService;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        routeSnapshot: ActivatedRoute
    ) {
        this.categoriesService = createClient(CategoriesService, this.transport);

        this.category.id = +routeSnapshot.snapshot.params['id'];
    }

    async ngOnInit() {
        try {
            let response = await this.categoriesService.listCategories({ ids: [+this.category.id] });
            if (response.categories && response.categories.length == 0) {
                this.messageService.add({ severity: 'error', detail: 'category not found' });
                return;
            }

            this.category = response.categories[0] ?? create(CategorySchema, {});
        } catch (e) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
        }
    }

    getTableTitle() {
        return `Transactions for category "${this.category.name}"`;
    }

    getFilter(): FilterWrapper {
        return {
            filters: {
                categories: {
                    matchMode: 'in',
                    value: [this.category.id]
                }
            }
        };
    }
}
