import { bootstrapApplication } from '@angular/platform-browser';
import { appConfig } from './app.config';
import { AppComponent } from './app.component';

declare global {
    interface BigInt {
        toJSON(): Number;
    }
}

BigInt.prototype.toJSON = function () { return Number(this) }

bootstrapApplication(AppComponent, appConfig).catch((err) => console.error(err));
