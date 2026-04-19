import { Inject, Injectable } from '@angular/core';
import { createClient, Transport } from '@connectrpc/connect';
import { ConfigurationService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/configuration/v1/configuration_pb';
import { TRANSPORT_TOKEN } from '../consts/transport';

@Injectable({ providedIn: 'root' })
export class PageConfigService {
    private readonly configService;

    constructor(@Inject(TRANSPORT_TOKEN) transport: Transport) {
        this.configService = createClient(ConfigurationService, transport);
    }

    async get<T>(pageId: string, defaults: T): Promise<T> {
        const key = this.buildKey(pageId);
        const response = await this.configService.getConfigsByKeys({ keys: [key] });
        const raw = response.configs[key];
        if (!raw) {
            return defaults;
        }
        try {
            const parsed = JSON.parse(raw) as Partial<T>;
            return { ...defaults, ...parsed };
        } catch (e) {
            console.error(`PageConfigService: failed to parse config for ${pageId}`, e);
            return defaults;
        }
    }

    async set<T>(pageId: string, value: T): Promise<void> {
        const key = this.buildKey(pageId);
        await this.configService.setConfigByKey({
            key,
            value: JSON.stringify(value),
        });
    }

    private buildKey(pageId: string): string {
        return `page.${pageId}`;
    }
}
