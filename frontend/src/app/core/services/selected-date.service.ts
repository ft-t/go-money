import { Injectable } from '@angular/core';
import { CookieService } from '../../services/cookie.service';
import { CookieInstances } from '../../objects/cookie-instances';
import { BehaviorSubject } from 'rxjs';

@Injectable()
export class SelectedDateService {
    private readonly FROM_DATE_COOKIE_NAME = 'go_money_from_date';
    private readonly TO_DATE_COOKIE_NAME = 'go_money_to_date';

    public fromDate = new BehaviorSubject<Date>(new Date());
    public toDate = new BehaviorSubject<Date>(new Date());

    constructor(private cookieService: CookieService) {
        this.fromDate.next(this.loadFromDate());
        this.toDate.next(this.loadToDate());
    }

    public loadFromDate(): Date {
        let val = this.cookieService.get(this.FROM_DATE_COOKIE_NAME);

        if (!val) {
            let targetDate = this.getFirstDayOfMonth();
            this.setFromDate(targetDate);

            return targetDate;
        }

        let parsedDate = new Date(val);

        if (!parsedDate) {
            console.error(`Invalid from date from cookie: ${val}`);

            let targetDate = this.getFirstDayOfMonth();
            this.setFromDate(targetDate);

            return targetDate;
        }

        return parsedDate;
    }

    public loadToDate(): Date {
        let val = this.cookieService.get(this.TO_DATE_COOKIE_NAME);

        if (!val) {
            let targetDate = this.getLastDayOfMonth();
            this.setToDate(targetDate);

            return targetDate;
        }

        let parsedDate = new Date(val);

        if (!parsedDate) {
            console.error(`Invalid to date from cookie: ${val}`);

            let targetDate = this.getLastDayOfMonth();
            this.setToDate(targetDate);

            return targetDate;
        }

        return parsedDate;
    }

    public getFirstDayOfMonth(): Date {
        let date = new Date();
        return new Date(date.getFullYear(), date.getMonth(), 1);
    }

    public getLastDayOfMonth(): Date {
        let date = new Date();
        return new Date(date.getFullYear(), date.getMonth() + 1, 0);
    }

    public setFromDate(date: Date): void {
        this.cookieService.set(this.FROM_DATE_COOKIE_NAME, date.toISOString(), {
            path: '/'
        });

        this.fromDate.next(date)
    }

    public setToDate(date: Date): void {
        this.cookieService.set(this.TO_DATE_COOKIE_NAME, date.toISOString(), {
            path: '/'
        });

        this.toDate.next(date)
    }
}
