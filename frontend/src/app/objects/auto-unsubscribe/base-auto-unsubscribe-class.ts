import {
  BehaviorSubject,
  finalize,
  MonoTypeOperatorFunction,
  of,
  OperatorFunction,
  Subject,
  Subscriber,
  Subscription
} from 'rxjs';
import { AutoUnsubscribe } from './auto-unsubscribe.decorator';
import { AutoUnsubscribeHelper } from './auto-unsubscribe.helper';
import { ObjectDictionary } from '../dictionary/object-dictionary';
import { DictionarySubscriptions } from '../dictionary/dictionary-subscriptions';
import { repeatSubscription } from '../repeat-subscription/repeat-subscription';
import { catchError, takeUntil } from 'rxjs/operators';

/**
 * Do not use this class without AutoUnsubscribe decorator if you don't have special necessaries
 * in some cases ivy incorrectly override component ɵfac factory on compile time and it will fail in runtime
 * when component try to use their own ɵfac property
 */
export class BaseAutoUnsubscribeClassUndecorated<T = string> {
  public cancellableSubject$: Subject<boolean> = new Subject<boolean>();

  public get takeUntilDestroy(): MonoTypeOperatorFunction<any> {
    return takeUntil(this.cancellableSubject$);
  }

  public catchDefault(methodName = ''): OperatorFunction<any, any> {
    return catchError((e: any) => {
      // console.error(methodName, 'default catch', e.message);
      return of(undefined);
    });
  }

  /**
   * used in repeatSubscription
   */
  public repeatSubscriptionsDictionary: ObjectDictionary<Subscriber<any>> = new ObjectDictionary<Subscriber<any>>();
  /**
   * could be used in templates
   */
  public repeatSubscription = repeatSubscription;
  protected subscriptions: Subscription[] = [];
  protected subscriptionsDictionary: DictionarySubscriptions<T> = new DictionarySubscriptions<T>();
  /**
   * used for not to unsubscribe twice, don't use outside
   */
    // tslint:disable-next-line:variable-name
  private _alreadyUnsubscribedFromAllPossibleSubscriptions = false;

  public _ngOnInit(): void {
    if (!this.cancellableSubject$ || this.cancellableSubject$.closed || this.cancellableSubject$.isStopped) {
      this.cancellableSubject$ = new Subject<boolean>();
    }
  }

  public _ngOnDestroy(): void {
    if (this._alreadyUnsubscribedFromAllPossibleSubscriptions) {
      return;
    }

    if (!this.cancellableSubject$.closed) {
      this.cancellableSubject$.next(true);
    }

    AutoUnsubscribeHelper.safeUnsubscribeFromAllProps(this);
    this._alreadyUnsubscribedFromAllPossibleSubscriptions = true;
  }

  public destroySubscriptionsPropsOnly(): void {
    AutoUnsubscribeHelper.safeUnsubscribeFromArray(this, 'subscriptions');
    this.subscriptionsDictionary.unsubscribe(AutoUnsubscribeHelper.safeUnsubscribe);
  }
}

/**
 * The logic in AutoUnsubscribe decorator is the same as here and safely reused but for some cases
 * decorator is still necessary because of some features of ɵfac work on compile time in some special cases
 */
@AutoUnsubscribe()
// @ts-ignore
export class BaseAutoUnsubscribeClass<T = string> extends BaseAutoUnsubscribeClassUndecorated<T> {
  public readonly isLoading$ = new BehaviorSubject<boolean>(false);

  // tslint:disable-next-line:use-lifecycle-interface
  ngOnInit(): void {
    super._ngOnInit();
  }

  // tslint:disable-next-line:use-lifecycle-interface
  ngOnDestroy(): void {
    super._ngOnDestroy();
  }

  public get finalizeLoading(): MonoTypeOperatorFunction<any> {
    return finalize(() => this.isLoading$.next(false));
  }
}
