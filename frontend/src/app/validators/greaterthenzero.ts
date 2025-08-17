import { AbstractControl, ValidationErrors, ValidatorFn } from '@angular/forms';

export const greaterThanZeroValidator =
    (): ValidatorFn =>
    (control: AbstractControl): ValidationErrors | null => {
        const value = Number(control.value);
        return value > 0 ? null : { greaterThanZero: true };
    };
