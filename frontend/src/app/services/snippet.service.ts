import { Injectable, signal, computed, Inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';

import { Snippet, SnippetTransaction } from '../models/snippet.model';

@Injectable({
    providedIn: 'root'
})
export class SnippetService {
    private readonly STORAGE_KEY = 'go_money_transaction_snippets';
    private readonly isBrowser: boolean;

    private snippetsSignal = signal<Snippet[]>([]);

    public snippets = computed(() => this.snippetsSignal());

    constructor(@Inject(PLATFORM_ID) platformId: Object) {
        this.isBrowser = isPlatformBrowser(platformId);
        this.loadFromStorage();
    }

    private loadFromStorage(): void {
        if (!this.isBrowser) return;

        try {
            const stored = localStorage.getItem(this.STORAGE_KEY);
            if (stored) {
                const parsed = JSON.parse(stored) as Snippet[];
                this.snippetsSignal.set(parsed);
            }
        } catch (e) {
            console.error('Failed to load snippets from localStorage:', e);
            this.snippetsSignal.set([]);
        }
    }

    private saveToStorage(): void {
        if (!this.isBrowser) return;

        try {
            localStorage.setItem(this.STORAGE_KEY, JSON.stringify(this.snippetsSignal()));
        } catch (e) {
            console.error('Failed to save snippets to localStorage:', e);
        }
    }

    private generateId(): string {
        return crypto.randomUUID();
    }

    createSnippet(name: string, transactions: SnippetTransaction[]): Snippet {
        const now = new Date().toISOString();
        const snippet: Snippet = {
            id: this.generateId(),
            name,
            transactions,
            createdAt: now,
            updatedAt: now
        };

        this.snippetsSignal.update(current => [...current, snippet]);
        this.saveToStorage();

        return snippet;
    }

    updateSnippet(id: string, transactions: SnippetTransaction[]): void {
        this.snippetsSignal.update(current =>
            current.map(s =>
                s.id === id
                    ? { ...s, transactions, updatedAt: new Date().toISOString() }
                    : s
            )
        );
        this.saveToStorage();
    }

    renameSnippet(id: string, newName: string): void {
        this.snippetsSignal.update(current =>
            current.map(s =>
                s.id === id
                    ? { ...s, name: newName, updatedAt: new Date().toISOString() }
                    : s
            )
        );
        this.saveToStorage();
    }

    deleteSnippet(id: string): void {
        this.snippetsSignal.update(current => current.filter(s => s.id !== id));
        this.saveToStorage();
    }

    getSnippetById(id: string): Snippet | undefined {
        return this.snippetsSignal().find(s => s.id === id);
    }
}
