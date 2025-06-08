import { Subscriber } from 'rxjs';
import { ObjectDictionary } from '../dictionary/object-dictionary';

export interface IRepeatSubscription<T = any> {
  repeatSubscriptionsDictionary: ObjectDictionary<Subscriber<T>>;
}
