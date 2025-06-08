import { ButtonModule } from 'primeng/button';
import { CheckboxModule } from 'primeng/checkbox';
import { InputTextModule } from 'primeng/inputtext';
import { PasswordModule } from 'primeng/password';
import { FormsModule } from '@angular/forms';
import { RippleModule } from 'primeng/ripple';
import { NgModule } from '@angular/core';
import { RouterModule } from '@angular/router';
import { AccountListComponent } from './account-list.component';
import { AccountListRoutingModule } from './account-list-routing.module';
import { UsersGrpcService } from '../../../services/auth/users-grpc.service';
import { TableModule } from 'primeng/table';
import { IconField } from 'primeng/iconfield';
import { InputIcon } from 'primeng/inputicon';
import { MultiSelect } from 'primeng/multiselect';
import { Select } from 'primeng/select';
import { Slider } from 'primeng/slider';
import { AsyncPipe, CurrencyPipe, DatePipe } from '@angular/common';
import { Tag } from 'primeng/tag';
import { ProgressBar } from 'primeng/progressbar';
import { CustomerService } from '../../../pages/service/customer.service';
import { ProductService } from '../../../pages/service/product.service';

@NgModule({
  imports: [
    ButtonModule,
    CheckboxModule,
    InputTextModule,
    PasswordModule,
    FormsModule,
    RouterModule,
    RippleModule,
    AccountListRoutingModule,
    TableModule,
    IconField,
    InputIcon,
    MultiSelect,
    Select,
    Slider,
    DatePipe,
    CurrencyPipe,
    Tag,
    ProgressBar,
    AsyncPipe
  ],
  declarations: [
    AccountListComponent
  ],
  exports: [
    AccountListComponent
  ],
  providers: [
    UsersGrpcService,
    CustomerService,
    ProductService
  ]
})
export class AccountListModule {}
