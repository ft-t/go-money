import {
  ChangeDetectionStrategy,
  Component,
  ElementRef,
  OnDestroy,
  OnInit,
  ViewChild
} from '@angular/core';
import { BaseAutoUnsubscribeClass } from '../../../objects/auto-unsubscribe/base-auto-unsubscribe-class';
import { Customer, CustomerService, Representative } from '../../../pages/service/customer.service';
import { Product, ProductService } from '../../../pages/service/product.service';
import { Table } from 'primeng/table';
import { BehaviorSubject } from 'rxjs';
import { AccountsGrpcService } from '../../../services/accounts/accounts-grpc.service';
import {
  ListAccountsResponse
} from '../../../../../gen/gomoneypb/accounts/v1/accounts_pb';
import { tap } from 'rxjs/operators';

@Component({
  selector: 'app-account-list',
  standalone: false,
  templateUrl: 'account-list.component.html',
  // styleUrls: ['account-list.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AccountListComponent extends BaseAutoUnsubscribeClass implements OnInit, OnDestroy {
  customers1$ = new BehaviorSubject<Customer[] | any>([])
  customers2: Customer[] = [];
  customers3: Customer[] = [];
  representatives: Representative[] = [];
  statuses: any[] = [];
  products: Product[] = [];
  activityValues: number[] = [0, 100];
  loading: boolean = false;

  @ViewChild('filter') filter!: ElementRef;

  constructor(private customerService: CustomerService,
              private accountsService: AccountsGrpcService,
              private productService: ProductService) {
    super();
  }

  override ngOnInit() {
    super.ngOnInit();

    this.customerService.getCustomersLarge().then((customers) => {
      this.customers1$.next(customers);
      this.loading = false;

      console.log(this.customers1$);

      // @ts-ignore
      this.customers1$.forEach((customer) => (customer.date = new Date(customer.date)));
    });
    this.customerService.getCustomersMedium().then((customers) => (this.customers2 = customers));
    this.customerService.getCustomersLarge().then((customers) => (this.customers3 = customers));
    this.productService.getProductsWithOrdersSmall().then((data) => (this.products = data));

    this.representatives = [
      {name: 'Amy Elsner', image: 'amyelsner.png'},
      {name: 'Anna Fali', image: 'annafali.png'},
      {name: 'Asiya Javayant', image: 'asiyajavayant.png'},
      {name: 'Bernardo Dominic', image: 'bernardodominic.png'},
      {name: 'Elwin Sharvill', image: 'elwinsharvill.png'},
      {name: 'Ioni Bowcher', image: 'ionibowcher.png'},
      {name: 'Ivan Magalhaes', image: 'ivanmagalhaes.png'},
      {name: 'Onyama Limba', image: 'onyamalimba.png'},
      {name: 'Stephen Shaw', image: 'stephenshaw.png'},
      {name: 'XuXue Feng', image: 'xuxuefeng.png'}
    ];

    this.statuses = [
      {label: 'Unqualified', value: 'unqualified'},
      {label: 'Qualified', value: 'qualified'},
      {label: 'New', value: 'new'},
      {label: 'Negotiation', value: 'negotiation'},
      {label: 'Renewal', value: 'renewal'},
      {label: 'Proposal', value: 'proposal'}
    ];
  }

  override ngOnDestroy() {
    super.ngOnDestroy();
  }

  getSeverity(status: string) {
    switch (status) {
      case 'qualified':
      case 'instock':
      case 'INSTOCK':
      case 'DELIVERED':
      case 'delivered':
        return 'success';

      case 'negotiation':
      case 'lowstock':
      case 'LOWSTOCK':
      case 'PENDING':
      case 'pending':
        return 'warn';

      case 'unqualified':
      case 'outofstock':
      case 'OUTOFSTOCK':
      case 'CANCELLED':
      case 'cancelled':
        return 'danger';

      default:
        return 'info';
    }
  }

  onGlobalFilter(table: Table, event: Event) {
    table.filterGlobal((event.target as HTMLInputElement).value, 'contains');
  }

  clear(table: Table) {
    table.clear();
    this.filter.nativeElement.value = '';
  }

  getClientsList() {
    if (this.isLoading$.value) {
      return;
    }

    this.isLoading$.next(true);

    // this.accountsService.listAccounts()
    //   .pipe(
    //     tap((response: ListAccountsResponse) => {
    //       console.log(response);
    //       // this.customers1$.next(response.accounts);
    //     }),
    //     this.takeUntilDestroy
    //   ).subscribe();
  }
}
