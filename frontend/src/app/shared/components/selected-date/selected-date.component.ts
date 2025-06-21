import { Component, OnInit, ViewEncapsulation } from '@angular/core';
import { SelectedDateService } from '../../../core/services/selected-date.service';
import { Button } from 'primeng/button';
import { DatePipe } from '@angular/common';
import { Popover } from 'primeng/popover';
import { Calendar } from 'primeng/calendar';
import { FormsModule } from '@angular/forms';
import { InputText } from 'primeng/inputtext';
import { InputMask } from 'primeng/inputmask';
import { DatePickerModule } from 'primeng/datepicker';

@Component({
    selector: 'selected-date',
    templateUrl: 'selected-date.component.html',
    imports: [Button, Popover, Calendar, FormsModule, DatePickerModule],
    styles: `
        customCalendar .p-datepicker-panel {
            position: relative !important;
        }
    `,
    encapsulation: ViewEncapsulation.None
})
export class SelectedDateComponent implements OnInit {
    public possibleDates: Date[][] = [];

    constructor(
        public selectedDateService: SelectedDateService,
        private datePipe: DatePipe
    ) {}

    public manualDateFrom: Date = new Date();
    public manualDateTo: Date = new Date();

    ngOnInit(): void {
        this.possibleDates = this.buildPreviousMonthDates();
        this.rangeDates = [this.selectedDateService.getFromDate(), this.selectedDateService.getToDate()];
        this.setManualDateFrom()
    }

    setManualDateFrom() {
        this.manualDateFrom = this.selectedDateService.getFromDate()
        this.manualDateTo = this.selectedDateService.getToDate()
    }

    rangeDates: Date[] = [];

    getCurrentFancyDate(): string {
        let from = this.selectedDateService.getFromDate();
        let to = this.selectedDateService.getToDate();

        return this.getFancyDate(from, to);
    }

    setDateFromButton(from: Date, to: Date) {
        this.manualDateTo = to;
        this.manualDateFrom = from;

       this.onManualDateChanged()
    }

    onManualDateChanged() {
        this.rangeDates = [this.manualDateFrom, this.manualDateTo];
        this.setDate()
    }

    onBigCalendarDateChanged() {
        let copy = this.rangeDates.slice();
        copy.sort()

        this.manualDateFrom = copy[0];
        this.manualDateTo = copy[1];

        this.setDate();
    }

    setDate() {
        this.selectedDateService.setFromDate(this.manualDateFrom);
        this.selectedDateService.setToDate(this.manualDateTo);

        this.possibleDates = this.buildPreviousMonthDates();
    }

    buildPreviousMonthDates() {
        let from = this.selectedDateService.getFromDate();
        let to = this.selectedDateService.getToDate();

        let firstDayOfCurrentMonth = this.selectedDateService.getFirstDayOfMonth();
        let possibleDates: Date[][] = [
            [firstDayOfCurrentMonth, this.selectedDateService.getLastDayOfMonth()]
        ];

        let max = 6;

        for (let i = 0; i < max; i++) {
            let safeFrom = new Date(from.getFullYear(), from.getMonth(), 1);
            let safeTo = new Date(to.getFullYear(), to.getMonth(), 1);

            safeFrom.setMonth(safeFrom.getMonth() - i);
            safeTo.setMonth(safeTo.getMonth() - i);

            if (safeFrom.toISOString() == firstDayOfCurrentMonth.toISOString()){
                max +=1;
                continue;
            }

            let copiedFromDate = new Date(safeFrom);
            copiedFromDate.setDate(Math.min(from.getDate(), this.daysInMonth(safeFrom)));

            let copiedToDate = new Date(safeTo);
            copiedToDate.setDate(Math.min(to.getDate(), this.daysInMonth(safeTo)));

            possibleDates.push([copiedFromDate, copiedToDate]);
        }

        return possibleDates;
    }

    daysInMonth(d: Date): number {
        return new Date(d.getFullYear(), d.getMonth() + 1, 0).getDate();
    }

    getFancyDate(from: Date, to: Date): string {
        return `${this.datePipe.transform(from, 'yyyy-MM-dd')} -
        ${this.datePipe.transform(to, 'yyyy-MM-dd')}`;
    }
}
