import { Operator, Subscriber, TeardownLogic } from 'rxjs';
import { IRepeatSubscription } from './i-repeat-subscription';
import { ObjectDictionary } from '../dictionary/object-dictionary';

export class RepeatSubscriptionOperator<T> implements Operator<T, T> {
  constructor(private dictionaryOrObjWithDictionary: IRepeatSubscription<T> | ObjectDictionary<Subscriber<T>>,
              private subscriptionDictionaryKey: string) {
  }

  call(subscriber: Subscriber<T>, source: any): TeardownLogic {
    const dictionary = this.dictionaryOrObjWithDictionary instanceof ObjectDictionary ? this.dictionaryOrObjWithDictionary
      : this.dictionaryOrObjWithDictionary.repeatSubscriptionsDictionary;

    const existedSubscriber = dictionary.getValue(this.subscriptionDictionaryKey);

    if (existedSubscriber && !existedSubscriber.closed) {
      existedSubscriber.complete();
    }

    dictionary.setValue(this.subscriptionDictionaryKey, subscriber);

    return source.subscribe(subscriber);
  }
}
