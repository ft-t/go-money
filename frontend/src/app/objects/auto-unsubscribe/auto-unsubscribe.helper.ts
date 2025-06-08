/* tslint:disable:forin variable-name ban-types max-line-length */
import { Subject, Subscription } from 'rxjs';
import { EventEmitter } from '@angular/core';
import { DictionarySubscriptions } from '../dictionary/dictionary-subscriptions';
import { PropertyHelper } from '../../helpers/property.helper';

interface IAutoUnsubscribeObserver {
  unsubscribe: Function;
  complete: Function;
  closed?: boolean;
  isStopped?: boolean;
}

type AutoUnsubscribeObserverType =
  IAutoUnsubscribeObserver
  | Omit<IAutoUnsubscribeObserver, 'unsubscribe'>
  | Omit<IAutoUnsubscribeObserver, 'complete'>;

export class AutoUnsubscribeHelper {
  public static safeUnsubscribe(observer: AutoUnsubscribeObserverType): void {
    const _observer = observer as IAutoUnsubscribeObserver;

    if (!_observer) {
      return;
    }

    const constructorName = _observer.constructor.name;

    if (constructorName === 'EventEmitter' && (_observer instanceof EventEmitter)) {
      return;
    }

    const hasUnsubscribe = PropertyHelper.isFunction(_observer.unsubscribe);
    const hasComplete = PropertyHelper.isFunction(_observer.complete);

    /**
     * we check for unsubscribe here, even in case we don't call it because it's have no sense to complete
     * observer globally that could not be unsubscribed, more important here to correctly complete their subscriptions
     */
    if (!hasUnsubscribe && !hasComplete) {
      return;
    }

    if (hasComplete && !_observer.closed) {
      _observer.complete();
    }

    if (!hasUnsubscribe) {
      return;
    }

    if (constructorName === 'Subscription' && (_observer instanceof Subscription)) {
      _observer.unsubscribe();
    } else if (constructorName === 'Subject' && (_observer instanceof Subject)) {
      _observer.unsubscribe();
    } else if (constructorName === 'DictionarySubscriptions' && (_observer instanceof DictionarySubscriptions)) {
      (_observer as DictionarySubscriptions).unsubscribe(AutoUnsubscribeHelper.safeUnsubscribe);
    } else if (!_observer.isStopped && !observer.closed) {
      _observer.unsubscribe();
    }
  }

  public static safeUnsubscribeFromArray(context: any, propName: string): void {
    if (!Array.isArray(context[propName])) {
      return;
    }

    (context[propName] as unknown as AutoUnsubscribeObserverType[]).forEach(AutoUnsubscribeHelper.safeUnsubscribe);
    (context[propName] as unknown as AutoUnsubscribeObserverType[]).length = 0;
  }

  public static safeUnsubscribeFromAllProps(context: any): void {
    for (const propName in context) {
      const property = context[propName];
      AutoUnsubscribeHelper.safeUnsubscribe(property);

      if (!Array.isArray(property)) {
        AutoUnsubscribeHelper.safeUnsubscribeFromArray(context, propName);
      }
    }
  }
}
