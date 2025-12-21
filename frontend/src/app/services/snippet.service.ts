import { Injectable, signal, computed, Inject } from '@angular/core';
import { createClient, Transport } from '@connectrpc/connect';

import { Transaction } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { ConfigurationService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/configuration/v1/configuration_pb';
import { Snippet, SnippetsConfig, serializeTransactions, deserializeTransactions } from '../models/snippet.model';
import { TRANSPORT_TOKEN } from '../consts/transport';

@Injectable({
    providedIn: 'root'
})
export class SnippetService {
    private readonly CONFIG_KEY = 'transaction_snippets';
    private readonly configService;

    private snippetsSignal = signal<Snippet[]>([]);
    private loadingSignal = signal<boolean>(false);
    private loadedSignal = signal<boolean>(false);

    public snippets = computed(() => this.snippetsSignal());
    public loading = computed(() => this.loadingSignal());

    constructor(@Inject(TRANSPORT_TOKEN) transport: Transport) {
        this.configService = createClient(ConfigurationService, transport);
    }

    async loadSnippets(): Promise<void> {
        if (this.loadedSignal()) return;

        this.loadingSignal.set(true);
        try {
            const response = await this.configService.getConfigsByKeys({
                keys: [this.CONFIG_KEY]
            });

            const configValue = response.configs[this.CONFIG_KEY];
            if (configValue) {
                const parsed = JSON.parse(configValue) as SnippetsConfig;
                this.snippetsSignal.set(parsed.snippets || []);
            } else {
                this.snippetsSignal.set([]);
            }
            this.loadedSignal.set(true);
        } catch (e) {
            console.error('Failed to load snippets:', e);
            this.snippetsSignal.set([]);
        } finally {
            this.loadingSignal.set(false);
        }
    }

    private async saveToBackend(): Promise<void> {
        const config: SnippetsConfig = {
            snippets: this.snippetsSignal()
        };

        try {
            await this.configService.setConfigByKey({
                key: this.CONFIG_KEY,
                value: JSON.stringify(config)
            });
        } catch (e) {
            console.error('Failed to save snippets:', e);
        }
    }

    private generateId(): string {
        return crypto.randomUUID();
    }

    async createSnippet(name: string, transactions: Transaction[]): Promise<Snippet> {
        const now = new Date().toISOString();
        const snippet: Snippet = {
            id: this.generateId(),
            name,
            transactions: serializeTransactions(transactions),
            createdAt: now,
            updatedAt: now
        };

        this.snippetsSignal.update(current => [...current, snippet]);
        await this.saveToBackend();

        return snippet;
    }

    async updateSnippet(id: string, transactions: Transaction[]): Promise<void> {
        this.snippetsSignal.update(current =>
            current.map(s =>
                s.id === id
                    ? { ...s, transactions: serializeTransactions(transactions), updatedAt: new Date().toISOString() }
                    : s
            )
        );
        await this.saveToBackend();
    }

    async renameSnippet(id: string, newName: string): Promise<void> {
        this.snippetsSignal.update(current =>
            current.map(s =>
                s.id === id
                    ? { ...s, name: newName, updatedAt: new Date().toISOString() }
                    : s
            )
        );
        await this.saveToBackend();
    }

    async deleteSnippet(id: string): Promise<void> {
        this.snippetsSignal.update(current => current.filter(s => s.id !== id));
        await this.saveToBackend();
    }

    async reorderSnippets(snippets: Snippet[]): Promise<void> {
        this.snippetsSignal.set(snippets);
        await this.saveToBackend();
    }

    getSnippetById(id: string): Snippet | undefined {
        return this.snippetsSignal().find(s => s.id === id);
    }

    getTransactions(snippet: Snippet): Transaction[] {
        return deserializeTransactions(snippet.transactions);
    }
}
