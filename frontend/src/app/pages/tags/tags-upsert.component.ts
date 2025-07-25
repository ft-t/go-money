import { Component, Inject, OnInit } from '@angular/core';
import { Fluid } from 'primeng/fluid';
import { InputText } from 'primeng/inputtext';
import { FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { MessageService } from 'primeng/api';
import { Tag, TagSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/tag_pb';
import { create } from '@bufbuild/protobuf';
import { Account, AccountSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/account_pb';
import { ErrorHelper } from '../../helpers/error.helper';
import { CreateTagRequestSchema, TagsService, UpdateTagRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/tags/v1/tags_pb';
import { ActivatedRoute, Router } from '@angular/router';
import { NgIf } from '@angular/common';
import { Button } from 'primeng/button';
import { Toast } from 'primeng/toast';
import { UpdateAccountRequestSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/accounts/v1/accounts_pb';
import { color } from 'chart.js/helpers';
import { ColorPickerModule } from 'primeng/colorpicker';
import { Message } from 'primeng/message';
import { CacheService } from '../../core/services/cache.service';

@Component({
    selector: 'app-tags-upsert',
    imports: [Fluid, InputText, ReactiveFormsModule, FormsModule, NgIf, Button, Toast, ColorPickerModule, Message],
    templateUrl: './tags-upsert.component.html'
})
export class TagsUpsertComponent implements OnInit {
    public tag: Tag = create(TagSchema, {
        color: '#11ff70'
    });
    private tagService;
    public form: FormGroup | undefined = undefined;

    constructor(
        @Inject(TRANSPORT_TOKEN) private transport: Transport,
        private messageService: MessageService,
        routeSnapshot: ActivatedRoute,
        private router: Router,
        private cache: CacheService
    ) {
        this.tagService = createClient(TagsService, this.transport);

        try {
            this.tag.id = +routeSnapshot.snapshot.params['id'];
        } catch (e) {
            this.tag.id = 0;
        }
    }

    async ngOnInit() {
        if (this.tag.id) {
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

        this.form = new FormGroup({
            id: new FormControl(this.tag.id, { nonNullable: false }),
            name: new FormControl(this.tag.name, Validators.required),
            icon: new FormControl(this.tag.icon),
            color: new FormControl(this.tag.color ?? '#a35050', Validators.required)
        });
    }

    async submit() {
        this.form!.markAllAsTouched();

        this.tag = this.form!.value as Tag;
        if (!this.form!.valid) {
            return;
        }

        this.cache.clear(); // tags svc

        if (this.tag.id) {
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
            let response = await this.tagService.updateTag(
                create(UpdateTagRequestSchema, {
                    name: this.tag.name,
                    icon: this.tag.icon,
                    color: this.tag.color,
                    id: this.tag.id
                })
            );

            this.messageService.add({ severity: 'info', detail: 'Tag updated' });
            await this.router.navigate(['/', 'tags', response.tag!.id.toString()]);
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    async create() {
        try {
            let response = await this.tagService.createTag(
                create(CreateTagRequestSchema, {
                    name: this.tag.name,
                    icon: this.tag.icon,
                    color: this.tag.color
                })
            );

            this.messageService.add({ severity: 'info', detail: 'Tag created' });
            await this.router.navigate(['/', 'tags', response.tag!.id.toString()]);
        } catch (e: any) {
            this.messageService.add({ severity: 'error', detail: ErrorHelper.getMessage(e) });
            return;
        }
    }

    protected readonly color = color;
}
