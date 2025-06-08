import { Injectable, OnDestroy } from '@angular/core';
import { Subject } from 'rxjs';
import { BaseAutoUnsubscribeClass } from '../objects/auto-unsubscribe/base-auto-unsubscribe-class';
import { ObjectDictionary } from '../objects/dictionary/object-dictionary';
import { AutoUnsubscribeHelper } from '../objects/auto-unsubscribe/auto-unsubscribe.helper';

export class BroadcastEvents {
  public static HandleCommonError = 'HandleCommonError';
  public static HandleConnectionLost = 'HandleConnectionLost';
  public static HandleUnexpectedErrors = 'HandleUnexpectedErrors';
  public static HandleJwtErrors = 'HandleJwtErrors';
  public static HandleRabbitError = 'HandleRabbitError';
  public static WindowScrolled = 'WindowScrolled';
  public static ReloadComponent = 'ReloadComponent';
  public static HandleCustomError = 'HandleCustomError';
  public static ReadAllMessages = 'ReadAllMessages';
}

@Injectable()
export class BroadcastService<U = BroadcastEvents> extends BaseAutoUnsubscribeClass implements OnDestroy {
  private listeners = new ObjectDictionary<Subject<any>>();

  // @ts-ignore
  public ngOnDestroy(): void {
    this.listeners.values().forEach(AutoUnsubscribeHelper.safeUnsubscribe);
    super.ngOnDestroy();
  }

  public emit<T>(type: string, data: T): void {
    const current = this.listeners.getValue(type);

    if (!current) {
      return;
    }

    current.next(data);
  }

  public listen<T>(type: string): Subject<T> {
    const current = this.listeners.getValue(type);

    if (current) {
      return current;
    }

    const emitter = new Subject<T>();
    this.listeners.setValue(type, emitter);

    return emitter;
  }
}
