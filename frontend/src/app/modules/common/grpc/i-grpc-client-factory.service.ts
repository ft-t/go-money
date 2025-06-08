import { BroadcastService } from '../../../services/broadcast.service';
import { PromiseClient } from '@connectrpc/connect';

export interface IGrpcClientFactoryService {

  broadcastService: BroadcastService;
  createClient(host: string, service: any): PromiseClient<any>;

  getJwt(): string;

  getXSeed(): string;

  hasDebug(): boolean;
}
