import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TableModule } from 'primeng/table';
import { ButtonDirective, ButtonLabel } from 'primeng/button';
import { IconField } from 'primeng/iconfield';
import { InputIcon } from 'primeng/inputicon';
import { InputText } from 'primeng/inputtext';

@NgModule({
  imports: [
    CommonModule,
    TableModule,
    ButtonDirective,
    ButtonLabel,
    IconField,
    InputIcon,
    InputText,
  ],
})
export class CurrencyListModule {}
