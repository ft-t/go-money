import { GrpcConfiguration } from './grpc.configuration';
import { from, Observable, of } from 'rxjs';
import { catchError, map, tap } from 'rxjs/operators';
import { Code, PromiseClient } from '@connectrpc/connect';
import { ServiceType } from '@bufbuild/protobuf';
import { IGrpcClientFactoryService } from '../i-grpc-client-factory.service';
import { PropertyHelper } from '../../../../helpers/property.helper';
import { CookieInstances } from '../../../../objects/cookie-instances';
import { CustomError } from './responses/custom-error';
import { CustomEnumError } from './responses/custom-enum-error';
import { BroadcastEvents } from '../../../../services/broadcast.service';

export class BaseGrpcConnectService {
  private readonly client: PromiseClient<any>;

  constructor(private grpcClientFactoryService: IGrpcClientFactoryService,
              grpcConfig: GrpcConfiguration,
              private service: any) {
    this.client = this.grpcClientFactoryService.createClient(grpcConfig.Host, service);
  }

  public getClient<T extends ServiceType>(): PromiseClient<T> {
    return this.client as PromiseClient<T>;
  }

  public send(methodName: string,
              params: any,
              jwtNeed: boolean = true,
              notPascalize: boolean = false,
              handleError: boolean = true,
              // @ts-ignore
              jwt: string = null): Observable<any> {
    const toLowCase = (str: string) => {
      return str.charAt(0).toLowerCase() + str.slice(1);
    };

    if (!params) {
      params = {};
    }

    if (!params.partnerId) {
      params.partnerId = 1;
    }

    let headers = new Headers();


    if (jwtNeed) {
      this.setJwtHeader(headers, jwt);
    }

    const res = this.client[`${toLowCase(methodName)}`]({
      ...params
    }, {
      headers
    });

    if (this.grpcClientFactoryService.hasDebug()) {
      console.log(`[REQUEST] ${toLowCase(methodName)}`, params);
    }

    return from(res).pipe(
      tap(r => {
        if (this.grpcClientFactoryService.hasDebug()) {
          console.log(`[RESPONSE] ${toLowCase(methodName)}`, r);
        }
      }),
      map((r: any) => {
        return notPascalize ? r : PropertyHelper.pascalizeObject(r);
      }),
      catchError(e => {
        console.error(methodName, e);

        try {
          const error = this.checkAndHandleError(e, handleError);

          if (handleError) {
            return of(error);
          }

          throw error;
        } catch (ex: any) {
          const CustomError = {
            ErrorCode: ex.code,
            ErrorMessage: 'error_something_went_wrong',
            OriginalError: ex.ErrorMessage,
            Data: null,
            IsTranslationNeed: true
          };

          if (handleError) {
            return of(CustomError);
          }

          throw CustomError;
        }
      })
    );
  }

  // @ts-ignore
  private setJwtHeader(header: Headers, customJwt: string = null): Headers {
    let jwt = null;

    if (customJwt) {
      jwt = customJwt;
    }

    if (!jwt && header.has(CookieInstances.Jwt)) {
      jwt = header.get(CookieInstances.Jwt);
    }

    if (!jwt) {
      jwt = this.grpcClientFactoryService.getJwt();
    }

    if (!jwt) {
      jwt = localStorage.getItem(CookieInstances.Jwt);
    }

    if (jwt) {
      header.set('Authorization', `${jwt}`);
    }

    return header;
  }

  private checkAndHandleError(response: any, handleError: boolean = true): CustomError {
    if (response.code === Code.Unauthenticated || response.statusMessage === 'client auth') {

      const customError = new CustomError({
        ErrorCode: CustomEnumError.MissingJwtToken,
        ErrorMessage: 'to_perform_this_action_you_need_to_be_authorized',
        OriginalError: response.rawMessage,
        Data: null,
        IsTranslationNeed: true
      });

      this.grpcClientFactoryService.broadcastService.emit(BroadcastEvents.HandleJwtErrors, CustomError);

      if (handleError) {
        this.grpcClientFactoryService.broadcastService.emit(BroadcastEvents.HandleCustomError, CustomError);
      }

      return customError;
    }

    switch (response.code) {

      default: {
        const customError = new CustomError({
          ErrorCode: response.status,
          ErrorMessage: 'error_something_went_wrong',
          OriginalError: response.rawMessage,
          Data: null,
          IsTranslationNeed: true
        });


        if (handleError) {
          this.grpcClientFactoryService.broadcastService.emit(BroadcastEvents.HandleCustomError, CustomError);
        }

        return customError;
      }
    }
  }
}
