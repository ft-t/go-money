import { Component, Inject, OnInit } from '@angular/core';
import { Fluid } from 'primeng/fluid';
import { InputText } from 'primeng/inputtext';
import { FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { MessageService } from 'primeng/api';
import { create } from '@bufbuild/protobuf';
import { ErrorHelper } from '../../helpers/error.helper';
import { ActivatedRoute, Router } from '@angular/router';
import { NgIf } from '@angular/common';
import { Button } from 'primeng/button';
import { Toast } from 'primeng/toast';
import { color } from 'chart.js/helpers';
import { ColorPickerModule } from 'primeng/colorpicker';
import { Message } from 'primeng/message';
import { Category, CategorySchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/category_pb';
import { CategoriesService, CreateCategoryRequestSchema, UpdateCategoryRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/categories/v1/categories_pb';
import { DefaultCache, ShortLivedCache } from '../../core/services/cache.service';

@Component({
    selector: 'app-categories-upsert',
    imports: [Fluid, InputText, ReactiveFormsModule, FormsModule, NgIf, Button, Toast, ColorPickerModule, Message],
    templateUrl: './categories-upsert.component.html'
})
export class CategoriesUpsertComponent implements OnInit {
    public category: Category = create(CategorySchema, {});
    private categoriesService;
    public form: FormGroup | undefined = undefined;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        routeSnapshot: ActivatedRoute,
        private router: Router,
        private defaultCache: DefaultCache,
        private shortLivedCache: ShortLivedCache
    ) {
        this.categoriesService = createClient(CategoriesService, this.transport);

        try {
            this.category.id = +routeSnapshot.snapshot.params['id'];
        } catch (e) {
            this.category.id = 0;
        }
    }

    async ngOnInit() {
        if (this.category.id) {
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

        this.form = new FormGroup({
            id: new FormControl(this.category.id, { nonNullable: false }),
            name: new FormControl(this.category.name, Validators.required)
        });
    }

    async submit() {
        this.form!.markAllAsTouched();

        this.category = this.form!.value as Category;
        if (!this.form!.valid) {
            return;
        }

        this.defaultCache.clear();
        this.shortLivedCache.clear();

        if (this.category.id) {
            await this.update();
        } else {
            await this.create();
        }
    }

    get name() {
        return this.form!.get('name')!;
    }

    async update() {
        try {
            let response = await this.categoriesService.updateCategory(
                create(UpdateCategoryRequestSchema, {
                    category: this.category
                })
            );

            this.messageService.add({ severity: 'info', detail: 'Category updated' });
            await this.router.navigate(['/', 'categories', response.category!.id.toString()]);
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    async create() {
        try {
            let response = await this.categoriesService.createCategory(
                create(CreateCategoryRequestSchema, {
                    category: this.category
                })
            );

            this.messageService.add({ severity: 'info', detail: 'Category created' });
            await this.router.navigate(['/', 'categories', response.category!.id.toString()]);
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    protected readonly color = color;
}
