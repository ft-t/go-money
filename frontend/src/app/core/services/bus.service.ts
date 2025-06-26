import { Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';

@Injectable()
export class BusService {
    public currentAccountId: BehaviorSubject<number> = new BehaviorSubject<number>(0);
}
