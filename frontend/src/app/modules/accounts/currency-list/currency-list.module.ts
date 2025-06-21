import { NgModule } from '@angular/core';
import { CurrencyListComponent } from './currency-list.component';
import { CurrencyListRoutingModule } from './currency-list-routing.module';
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
    CurrencyListRoutingModule,
    TableModule,
    ButtonDirective,
    ButtonLabel,
    IconField,
    InputIcon,
    InputText,
    ReactiveComponentModule
  ],
  declarations: [
    CurrencyListComponent
  ],
  exports: [
    CurrencyListComponent
  ],
  providers: [
    ProductService
  ]
})
export class CurrencyListModule {}
