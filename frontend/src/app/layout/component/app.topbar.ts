import { Component } from '@angular/core';
import { MenuItem } from 'primeng/api';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { StyleClassModule } from 'primeng/styleclass';
import { AppConfigurator } from './app.configurator';
import { LayoutService } from '../service/layout.service';
import { CookieService } from '../../services/cookie.service';
import { CookieInstances } from '../../objects/cookie-instances';
import { SelectedDateComponent } from '../../shared/components/selected-date/selected-date.component';
import { InputText } from 'primeng/inputtext';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

@Component({
    selector: 'app-topbar',
    standalone: true,
    imports: [RouterModule, CommonModule, StyleClassModule, AppConfigurator, SelectedDateComponent, InputText, ReactiveFormsModule, FormsModule],
    template: ` <div class="layout-topbar flex items-center justify-between w-full">
        <div class="layout-topbar-logo-container flex items-center">
            <button class="layout-menu-button layout-topbar-action" (click)="layoutService.onMenuToggle()">
                <i class="pi pi-bars"></i>
            </button>
            <a class="layout-topbar-logo" routerLink="/">
                <img class="w-24 shrink-0 mx-auto" src="/logo.png" alt="logo" />
            </a>
        </div>

        <div class="flex-1 flex justify-center">
            <input (keyup.enter)="search()" type="text" pInputText [(ngModel)]="searchValue" placeholder="Search" class="w-full max-w-xs" />
        </div>

        <div class="layout-topbar-actions flex items-center">
            <div class="layout-config-menu">
                <button type="button" class="layout-topbar-action" (click)="toggleDarkMode()">
                    <i [ngClass]="{ 'pi ': true, 'pi-moon': layoutService.isDarkTheme(), 'pi-sun': !layoutService.isDarkTheme() }"></i>
                </button>
                <div class="relative">
                    <button
                        class="layout-topbar-action layout-topbar-action-highlight"
                        pStyleClass="@next"
                        enterFromClass="hidden"
                        enterActiveClass="animate-scalein"
                        leaveToClass="hidden"
                        leaveActiveClass="animate-fadeout"
                        [hideOnOutsideClick]="true"
                    >
                        <i class="pi pi-palette"></i>
                    </button>
                    <app-configurator />
                </div>
            </div>

            <button class="layout-topbar-menu-button layout-topbar-action" pStyleClass="@next" enterFromClass="hidden" enterActiveClass="animate-scalein" leaveToClass="hidden" leaveActiveClass="animate-fadeout" [hideOnOutsideClick]="true">
                <i class="pi pi-ellipsis-v"></i>
            </button>
            <selected-date />

            <div class="layout-topbar-menu hidden lg:block">
                <div class="layout-topbar-menu-content">
                    <button (click)="logout()" type="button" class="layout-topbar-action">
                        <i class="pi pi-user"></i>
                        <span>Profile</span>
                    </button>
                </div>
            </div>
        </div>
    </div>`
})
export class AppTopbar {
    items!: MenuItem[];
    searchValue: string = '';

    constructor(
        public layoutService: LayoutService,
        private cookieService: CookieService,
        private router: Router
    ) {}

    async search() {
        await this.router.navigate(['/', 'transactions'], {
            queryParams: {
                title: this.searchValue,
                ignoreDateFilter: true
            }
        });
    }

    toggleDarkMode() {
        this.layoutService.layoutConfig.update((state) => ({ ...state, darkTheme: !state.darkTheme }));
    }

    logout() {
        this.cookieService.delete(CookieInstances.Jwt);
        return this.router.navigate(['/', 'login']);
    }
}
