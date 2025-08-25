import { Component, Inject, OnInit } from '@angular/core';
import { TRANSPORT_TOKEN } from '../../consts/transport';
import { createClient, Transport } from '@connectrpc/connect';
import { ConfigurationService, GetConfigurationResponse, GetConfigurationResponseSchema } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/configuration/v1/configuration_pb';
import { create } from '@bufbuild/protobuf';

@Component({
    standalone: true,
    selector: 'app-footer',
    template: `
        <div class="layout-footer flex justify-end">
            <div class="flex flex-col">
                <a target="_blank" href="https://github.com/ft-t/go-money"> Go Money {{ config.backendVersion }}
                    <a [href]="getLink()" target="_blank">({{ config.commitSha }})</a>
                </a>
            </div>
        </div>`
})
export class AppFooter implements OnInit {
    private configService;
    public config: GetConfigurationResponse = create(GetConfigurationResponseSchema, {});

    constructor(@Inject(TRANSPORT_TOKEN) private transport: Transport) {
        this.configService = createClient(ConfigurationService, this.transport);
    }

    async ngOnInit() {
        this.config = await this.configService.getConfiguration({});
    }

    getLink() {
        return `https://github.com/ft-t/go-money/commit/${this.config.commitSha}`
    }
}
