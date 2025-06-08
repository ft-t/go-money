import { MonoTypeOperatorFunction, Observable, Subscriber } from 'rxjs';
import { RepeatSubscriptionOperator } from './repeat-subscription-operator';
import { IRepeatSubscription } from './i-repeat-subscription';
import { ObjectDictionary } from '../dictionary/object-dictionary';

/**
 * when you need to subscribe multiple times and unsubscribe from previous subscriptions,
 * you can use this operator with unique string key
 * @param dictionaryOrObjWithDictionary
 * @param subscriptionDictionaryKey
 */
export function repeatSubscription<T>(
  dictionaryOrObjWithDictionary: IRepeatSubscription<T> | ObjectDictionary<Subscriber<T>>,
  subscriptionDictionaryKey: string): MonoTypeOperatorFunction<T> {
  return (source: Observable<T>) =>
    source.lift(new RepeatSubscriptionOperator(dictionaryOrObjWithDictionary, subscriptionDictionaryKey));
}
