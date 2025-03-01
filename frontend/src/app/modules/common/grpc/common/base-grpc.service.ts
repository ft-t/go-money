import { GrpcConfiguration } from './grpc.configuration';
// import { grpc } from '@improbable-eng/grpc-web';
// import { GrpcMessage } from './grpc-message';
// import { NodeHttpTransport } from '@improbable-eng/grpc-web-node-http-transport';
import { Inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { Observable } from 'rxjs';
import { PropertyHelper } from '../../../../helpers/property.helper';

// const xhr = grpc.CrossBrowserHttpTransport({});
// const ws = grpc.WebsocketTransport();

export class BaseGrpcService {
  // tslint:disable-next-line:variable-name
  protected readonly _service: any;
  // tslint:disable-next-line:variable-name
  // @ts-ignore
  protected readonly _serviceUrl: string;
  // @ts-ignore
  protected readonly isBrowser: boolean;

  constructor(grpcConfig: GrpcConfiguration,
              @Inject(PLATFORM_ID) platformId: any) {
    // this._serviceUrl = grpcConfig.Service;
    // this.isBrowser = isPlatformBrowser(platformId);
    // this._service = grpcClientFactory.createClient({
    //   host: this.isBrowser ? grpcConfig.Host : grpcConfig.SSRHost,
    //   service: grpcConfig.Service,
    //   debug: true,
    //   transport: this.isBrowser ? {
    //     unary: xhr,
    //     serverStream: xhr,
    //     clientStream: ws,
    //     bidiStream: ws,
    //   } : NodeHttpTransport(),
    // });
  }

  public send<Q , S>(
    methodName: string,
    params: any,
    reqclss: any,
    resclss: any,
    jwtNeed: boolean = true,
    notPascalize: boolean = false,
    handleError: boolean = true
  ): Observable<any> {
    const req = new reqclss();

    Object.keys(params).forEach(key => {
      const setFunction = req[`set${PropertyHelper.pascalizeString(key)}`];

      if (!setFunction || typeof setFunction !== 'function') {
        console.warn('wrong key!', key, 'service:', this._serviceUrl);
        return;
      }

      setFunction.call(req, params[key]);
    });

    // @ts-ignore
    return this._service.unary<Q, S>(methodName,
      req,
      null,
      reqclss,
      resclss,
      this.isBrowser,
      jwtNeed,
      notPascalize,
      handleError
    );
  }

  // @ts-ignore
  public setArrayToRequest(items: any[], reqclss): any[] {
    if (!items || !items.length) {
      return [];
    }

    const res: any[] = [];

    items.forEach(item => {
      const req = new reqclss();
      Object.keys(item).forEach(key => {
        const setFunction = req[`set${PropertyHelper.pascalizeString(key)}`];

        if (key.endsWith('Map')) {
          this.setMapToRequest(item[key], key, req);

          return;
        }

        if (!setFunction || typeof setFunction !== 'function') {
          console.warn('wrong key!', key, 'service:', this._serviceUrl);
          return;
        }

        try {
          setFunction.call(req, item[key]);
        } catch (ex) {
          console.log('error with key', key);
          throw ex;
        }

      });

      res.push(req);
    });


    return res;
  }

  public setMapToRequest(mapToSet: any, key: string, req: any): any {
    const getMapFunction = req[`get${PropertyHelper.pascalizeString(key)}`];

    const map = getMapFunction.call(req);

    Object.keys(mapToSet).forEach(mapKey => {
      map.set(mapKey, mapToSet[mapKey]);
    });

    return req;
  }
}
