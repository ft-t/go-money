import { ObjectDictionary } from './object-dictionary';
import { Subscription } from 'rxjs';

export class DictionarySubscriptions<T = string> {
  public get rawStore(): ObjectDictionary<Subscription> {
    return this.store;
  }

  private store: ObjectDictionary<Subscription> = new ObjectDictionary<Subscription>();

  public setValue(subscriptionKey: T, value: Subscription): void {
    this.remove(subscriptionKey);

    this.store.setValue((subscriptionKey as unknown as string).toString(), value);
  }

  public getValue(subscriptionKey: T): Subscription | null {
    return this.store.getValue((subscriptionKey as unknown as string).toString());
  }

  public remove(subscriptionKey: T): void {
    const existedSubscription = this.store.getValue((subscriptionKey as unknown as string).toString());

    if (!existedSubscription) {
      return;
    }

    !existedSubscription.closed && existedSubscription.unsubscribe();
    this.store.remove((subscriptionKey as unknown as string).toString());
  }

  public removeAll(customUnsubscribe?: (item: Subscription) => void): void {
    if (customUnsubscribe) {
      this.store.values().forEach(customUnsubscribe);
    } else {
      this.store.values().forEach((item) => !item.closed && item.unsubscribe());
    }

    this.store.removeAll();
  }

  /**
   * unsubscribe from all
   * used in AutoUnsubscribe
   */
  public unsubscribe(customUnsubscribe?: (item: Subscription) => void): void {
    this.removeAll(customUnsubscribe);
  }
}
