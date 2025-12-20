import { Component, EventEmitter, HostListener, Input, Output, signal, inject, computed, ElementRef, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ButtonModule } from 'primeng/button';
import { InputTextModule } from 'primeng/inputtext';
import { TooltipModule } from 'primeng/tooltip';
import { IconField } from 'primeng/iconfield';
import { InputIcon } from 'primeng/inputicon';
import { ConfirmationService, MessageService } from 'primeng/api';
import { ConfirmDialogModule } from 'primeng/confirmdialog';

import { Transaction } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/v1/transaction_pb';
import { SnippetService } from '../../../services/snippet.service';
import { Snippet } from '../../../models/snippet.model';

@Component({
    selector: 'snippet-spotlight',
    templateUrl: 'snippet-spotlight.component.html',
    imports: [
        CommonModule,
        FormsModule,
        ButtonModule,
        InputTextModule,
        TooltipModule,
        ConfirmDialogModule,
        IconField,
        InputIcon
    ]
})
export class SnippetSpotlightComponent {
    private snippetService = inject(SnippetService);
    private confirmationService = inject(ConfirmationService);
    private messageService = inject(MessageService);

    @ViewChild('searchInput') searchInput!: ElementRef<HTMLInputElement>;

    @Input() visible = false;
    @Output() visibleChange = new EventEmitter<boolean>();
    @Output() applySnippet = new EventEmitter<Transaction[]>();
    @Input() currentTransactions: Transaction[] = [];

    filterText = signal('');
    editingSnippetId = signal<string | null>(null);
    editingName = signal<string>('');
    selectedIndex = signal(0);
    draggedSnippetId = signal<string | null>(null);
    dragOverSnippetId = signal<string | null>(null);

    filteredSnippets = computed(() => {
        const filter = this.filterText().toLowerCase().trim();
        const all = this.snippetService.snippets();
        if (!filter) return all;
        return all.filter(s => s.name.toLowerCase().includes(filter));
    });

    @HostListener('document:keydown', ['$event'])
    handleKeyboardEvent(event: KeyboardEvent) {
        if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
            event.preventDefault();
            this.toggle();
            return;
        }

        if (!this.visible) return;

        if (this.editingSnippetId()) return;

        switch (event.key) {
            case 'Escape':
                event.preventDefault();
                this.close();
                break;
            case 'ArrowDown':
                event.preventDefault();
                this.moveSelection(1);
                break;
            case 'ArrowUp':
                event.preventDefault();
                this.moveSelection(-1);
                break;
            case 'Enter':
                event.preventDefault();
                this.applySelected();
                break;
        }
    }

    private moveSelection(delta: number): void {
        const filtered = this.filteredSnippets();
        if (filtered.length === 0) return;

        let newIndex = this.selectedIndex() + delta;
        if (newIndex < 0) newIndex = filtered.length - 1;
        if (newIndex >= filtered.length) newIndex = 0;
        this.selectedIndex.set(newIndex);
    }

    private applySelected(): void {
        const filtered = this.filteredSnippets();
        const idx = this.selectedIndex();
        if (idx >= 0 && idx < filtered.length) {
            this.onApply(filtered[idx]);
        }
    }

    toggle(): void {
        this.visible = !this.visible;
        this.visibleChange.emit(this.visible);

        if (this.visible) {
            this.filterText.set('');
            this.selectedIndex.set(0);
            this.focusSearch();
        } else {
            this.cancelEditing();
        }
    }

    open(): void {
        this.visible = true;
        this.visibleChange.emit(true);
        this.filterText.set('');
        this.selectedIndex.set(0);
        this.focusSearch();
    }

    private focusSearch(): void {
        setTimeout(() => {
            this.searchInput?.nativeElement?.focus();
        }, 100);
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

    onFilterChange(value: string): void {
        this.filterText.set(value);
        this.selectedIndex.set(0);
    }

    onApply(snippet: Snippet): void {
        const transactions = this.snippetService.getTransactions(snippet);
        this.applySnippet.emit(transactions);
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

        const name = `Snippet ${this.snippetService.snippets().length + 1}`;
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

    isSelected(snippet: Snippet): boolean {
        const filtered = this.filteredSnippets();
        const idx = this.selectedIndex();
        return filtered[idx]?.id === snippet.id;
    }

    getTransactionCount(snippet: Snippet): number {
        return snippet.transactions.length;
    }

    onDragStart(event: DragEvent, snippet: Snippet): void {
        if (this.filterText()) return;
        this.draggedSnippetId.set(snippet.id);
        if (event.dataTransfer) {
            event.dataTransfer.effectAllowed = 'move';
            event.dataTransfer.setData('text/plain', snippet.id);
        }
    }

    onDragOver(event: DragEvent, snippet: Snippet): void {
        if (!this.draggedSnippetId() || this.filterText()) return;
        event.preventDefault();
        if (event.dataTransfer) {
            event.dataTransfer.dropEffect = 'move';
        }
        this.dragOverSnippetId.set(snippet.id);
    }

    onDragLeave(): void {
        this.dragOverSnippetId.set(null);
    }

    onDragEnd(): void {
        this.draggedSnippetId.set(null);
        this.dragOverSnippetId.set(null);
    }

    onDrop(event: DragEvent, targetSnippet: Snippet): void {
        event.preventDefault();
        const draggedId = this.draggedSnippetId();
        if (!draggedId || draggedId === targetSnippet.id || this.filterText()) {
            this.onDragEnd();
            return;
        }

        const snippets = [...this.snippetService.snippets()];
        const draggedIndex = snippets.findIndex(s => s.id === draggedId);
        const targetIndex = snippets.findIndex(s => s.id === targetSnippet.id);

        if (draggedIndex === -1 || targetIndex === -1) {
            this.onDragEnd();
            return;
        }

        const [removed] = snippets.splice(draggedIndex, 1);
        snippets.splice(targetIndex, 0, removed);

        this.snippetService.reorderSnippets(snippets);
        this.onDragEnd();
    }

    isDragging(snippet: Snippet): boolean {
        return this.draggedSnippetId() === snippet.id;
    }

    isDragOver(snippet: Snippet): boolean {
        return this.dragOverSnippetId() === snippet.id && this.draggedSnippetId() !== snippet.id;
    }

    canDrag(): boolean {
        return !this.filterText();
    }
}
