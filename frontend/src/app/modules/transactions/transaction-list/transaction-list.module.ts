import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ProductService } from '../../../pages/service/product.service';
import { TableModule } from 'primeng/table';
import { ButtonDirective, ButtonLabel } from 'primeng/button';
import { IconField } from 'primeng/iconfield';
import { InputIcon } from 'primeng/inputicon';
import { InputText } from 'primeng/inputtext';
import { ReactiveComponentModule } from '../../common/reactive-component';

@NgModule({
  imports: [
    CommonModule,
    TableModule,
    ButtonDirective,
    ButtonLabel,
    IconField,
    InputIcon,
    InputText,
    ReactiveComponentModule
  ],
  providers: [
    ProductService
  ]
})
export class CurrencyListModule {}
