import { Component } from '@angular/core';

@Component({
    standalone: true,
    selector: 'app-footer',
    template: `<div class="layout-footer flex justify-end">
        <div class="flex flex-col">
            <a target="_blank" href="https://github.com/ft-t/go-money">
                Go Money v{{ '0.1.0' }} alpha
            </a>
        </div>
    </div>`
})
export class AppFooter {}
