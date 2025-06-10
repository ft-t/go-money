import { Transport} from '@connectrpc/connect';
import { InjectionToken } from '@angular/core';

export const TRANSPORT_TOKEN = new InjectionToken<Transport>('Connect Transport');
