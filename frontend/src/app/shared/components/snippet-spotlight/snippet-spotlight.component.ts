import { Component, EventEmitter, HostListener, Input, Output, signal, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ButtonModule } from 'primeng/button';
import { InputTextModule } from 'primeng/inputtext';
import { Tooltip } from 'primeng/tooltip';
import { ConfirmationService, MessageService } from 'primeng/api';
import { ConfirmDialogModule } from 'primeng/confirmdialog';

import { SnippetService } from '../../../services/snippet.service';
import { Snippet, SnippetTransaction } from '../../../models/snippet.model';

@Component({
    selector: 'snippet-spotlight',
    templateUrl: 'snippet-spotlight.component.html',
    imports: [
        CommonModule,
        FormsModule,
        ButtonModule,
        InputTextModule,
        Tooltip,
        ConfirmDialogModule
    ]
})
export class SnippetSpotlightComponent {
    private snippetService = inject(SnippetService);
    private confirmationService = inject(ConfirmationService);
    private messageService = inject(MessageService);

    @Input() visible = false;
    @Output() visibleChange = new EventEmitter<boolean>();
    @Output() applySnippet = new EventEmitter<SnippetTransaction[]>();
    @Input() currentTransactions: SnippetTransaction[] = [];

    editingSnippetId = signal<string | null>(null);
    editingName = signal<string>('');

    get snippets() {
        return this.snippetService.snippets();
    }

    @HostListener('document:keydown', ['$event'])
    handleKeyboardEvent(event: KeyboardEvent) {
        if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
            event.preventDefault();
            this.toggle();
        }

        if (event.key === 'Escape' && this.visible) {
            this.close();
        }
    }

    toggle(): void {
        this.visible = !this.visible;
        this.visibleChange.emit(this.visible);

        if (!this.visible) {
            this.cancelEditing();
        }
    }

    open(): void {
        this.visible = true;
        this.visibleChange.emit(true);
    }

    close(): void {
        this.visible = false;
        this.visibleChange.emit(false);
        this.cancelEditing();
    }

    onBackdropClick(event: MouseEvent): void {
        if ((event.target as HTMLElement).classList.contains('spotlight-backdrop')) {
            this.close();
        }
    }

    onApply(snippet: Snippet): void {
        this.applySnippet.emit(snippet.transactions);
        this.close();
        this.messageService.add({
            severity: 'success',
            detail: `Applied snippet "${snippet.name}"`
        });
    }

    onCreateNew(): void {
        if (this.currentTransactions.length === 0) {
            this.messageService.add({
                severity: 'warn',
                detail: 'No transactions to save as snippet'
            });
            return;
        }

        const name = `Snippet ${this.snippets.length + 1}`;
        const created = this.snippetService.createSnippet(name, this.currentTransactions);

        this.messageService.add({
            severity: 'success',
            detail: `Created snippet "${created.name}"`
        });

        this.startEditing(created.id, created.name);
    }

    onUpdate(snippet: Snippet): void {
        if (this.currentTransactions.length === 0) {
            this.messageService.add({
                severity: 'warn',
                detail: 'No transactions to update snippet with'
            });
            return;
        }

        this.snippetService.updateSnippet(snippet.id, this.currentTransactions);

        this.messageService.add({
            severity: 'success',
            detail: `Updated snippet "${snippet.name}"`
        });
    }

    onDelete(snippet: Snippet): void {
        this.confirmationService.confirm({
            message: `Are you sure you want to delete "${snippet.name}"?`,
            header: 'Confirm Delete',
            icon: 'pi pi-exclamation-triangle',
            acceptButtonStyleClass: 'p-button-danger',
            accept: () => {
                this.snippetService.deleteSnippet(snippet.id);
                this.messageService.add({
                    severity: 'success',
                    detail: `Deleted snippet "${snippet.name}"`
                });
            }
        });
    }

    startEditing(id: string, currentName: string): void {
        this.editingSnippetId.set(id);
        this.editingName.set(currentName);
    }

    saveEditing(): void {
        const id = this.editingSnippetId();
        const name = this.editingName().trim();

        if (id && name) {
            this.snippetService.renameSnippet(id, name);
        }

        this.cancelEditing();
    }

    cancelEditing(): void {
        this.editingSnippetId.set(null);
        this.editingName.set('');
    }

    onEditKeydown(event: KeyboardEvent): void {
        event.stopPropagation();
        if (event.key === 'Enter') {
            event.preventDefault();
            this.saveEditing();
        }
        if (event.key === 'Escape') {
            event.preventDefault();
            this.cancelEditing();
        }
    }
}
