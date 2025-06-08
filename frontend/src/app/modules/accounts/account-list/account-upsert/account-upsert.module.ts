import { ButtonModule } from 'primeng/button';
import { CheckboxModule } from 'primeng/checkbox';
import { InputTextModule } from 'primeng/inputtext';
import { PasswordModule } from 'primeng/password';
import { FormsModule } from '@angular/forms';
import { RippleModule } from 'primeng/ripple';
import { NgModule } from '@angular/core';
import { RouterModule } from '@angular/router';
import { AccountUpsertComponent } from './account-upsert.component';
import { AccountUpsertRoutingModule } from './account-upsert-routing.module';
import { TableModule } from 'primeng/table';
import { IconField } from 'primeng/iconfield';
import { InputIcon } from 'primeng/inputicon';
import { MultiSelect } from 'primeng/multiselect';
import { Select } from 'primeng/select';
import { Slider } from 'primeng/slider';
import { AsyncPipe, CurrencyPipe, DatePipe } from '@angular/common';
import { Tag } from 'primeng/tag';
import { ProgressBar } from 'primeng/progressbar';
import { Fluid } from 'primeng/fluid';
import { Textarea } from 'primeng/textarea';

@NgModule({
  imports: [
    ButtonModule,
    CheckboxModule,
    InputTextModule,
    PasswordModule,
    FormsModule,
    RouterModule,
    RippleModule,
    AccountUpsertRoutingModule,
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
    AsyncPipe,
    Fluid,
    Textarea
  ],
  declarations: [
    AccountUpsertComponent
  ],
  exports: [
    AccountUpsertComponent
  ],
  providers: [
  ]
})
export class AccountUpsertModule {}
