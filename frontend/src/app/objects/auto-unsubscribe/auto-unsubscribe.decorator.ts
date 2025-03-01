/* tslint:disable:no-string-literal */
import { Subject } from 'rxjs';
import { AutoUnsubscribeHelper } from './auto-unsubscribe.helper';
import { PropertyHelper } from '../../helpers/property.helper';

/**
 * Class Decorator That unsubscribe from Subscriptions or Subscription array
 */
export function AutoUnsubscribe() {
  // tslint:disable-next-line:only-arrow-functions ban-types
  return function (constructor: Function) {
    const originalNgOnInit = constructor.prototype['ngOnInit'];
    const originalNgOnDestroy = constructor.prototype['ngOnDestroy'];

    constructor.prototype.cancellableSubject$ = new Subject<boolean>();

    constructor.prototype['ngOnInit'] = function () {
      const callOriginalNgOnInit = () => PropertyHelper.isFunction(originalNgOnInit) && originalNgOnInit.apply(this);

      if (!this.cancellableSubject$ || this.cancellableSubject$.closed || this.cancellableSubject$.isStopped) {
        this.cancellableSubject$ = new Subject<boolean>();
      }

      callOriginalNgOnInit();
    };

    constructor.prototype['ngOnDestroy'] = function () {
      const callOriginalNgOnDestroy = () => PropertyHelper.isFunction(originalNgOnDestroy) && originalNgOnDestroy.apply(this);

      if (this._alreadyUnsubscribedFromAllPossibleSubscriptions) {
        callOriginalNgOnDestroy();
        return;
      }

      // tslint:disable-next-line:no-unused-expression
      !this.cancellableSubject$.closed && this.cancellableSubject$.next(true);

      AutoUnsubscribeHelper.safeUnsubscribeFromAllProps(this);
      this._alreadyUnsubscribedFromAllPossibleSubscriptions = true;

      callOriginalNgOnDestroy();
    };
  };
}
