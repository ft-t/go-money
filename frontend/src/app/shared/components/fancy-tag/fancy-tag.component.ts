import { Component, Input } from '@angular/core';
import { Tag as Tag2, TagSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/tag_pb';
import { create } from '@bufbuild/protobuf';
import { Tag } from 'primeng/tag';
import { NgIf } from '@angular/common';

@Component({
    selector: 'fancy-tag',
    imports: [Tag, NgIf],
    templateUrl: './fancy-tag.component.html'
})
export class FancyTagComponent {
    @Input() public tag: Tag2 | undefined = create(TagSchema, {});
}
