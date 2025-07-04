import { Component, ElementRef } from '@angular/core';
import { AppMenu } from './app.menu';
import { FormsModule } from '@angular/forms';
import { InputText } from 'primeng/inputtext';

@Component({
    selector: 'app-sidebar',
    standalone: true,
    imports: [AppMenu, FormsModule],
    template: ` <div class="layout-sidebar">
        <app-menu></app-menu>
    </div>`
})
export class AppSidebar {
    constructor(public el: ElementRef) {}
}
