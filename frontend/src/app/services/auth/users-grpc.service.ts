import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { ConfigService } from '../config.service';
import { CookieService } from '../cookie.service';
// import { GrpcClientFactoryService } from '../../modules/common/grpc/grpc-client-factory.service';
// import { BaseGrpcConnectService } from '../../modules/common/grpc/common/base-grpc-connect.service';
import { Observable } from 'rxjs';
import { UsersService } from '../../../../gen/gomoneypb/users/v1/users_connect';
import { CreateRequest, CreateResponse, LoginRequest, LoginResponse } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/users/v1/users_pb';
import { CookieInstances } from '../../objects/cookie-instances';

// import { LoginRequest } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/users/v1/users_pb';

@Injectable({
    providedIn: 'root'
})
export class UsersGrpcService {
    private readonly router = inject(Router);

    constructor(
        private configurationService: ConfigService,
        private cookieService: CookieService,
        private httpClient: HttpClient
    ) {}

    public loginMethod(request: LoginRequest) {
        // return this.send('Login',
        //   {
        //     request
        //   },
        //   false
        // )
    }

    public register(request: CreateRequest) {
        //   return this.send('Create',
        //     {
        //       request
        //     },
        //     false
        //   )
    }

    public logout() {
        this.cookieService.delete(CookieInstances.Jwt);
        this.router.navigate(['/', 'login']);
    }
}
