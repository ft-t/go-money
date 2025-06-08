import { CanActivateFn, Router, UrlTree } from '@angular/router';
import { inject } from '@angular/core';
import { Observable } from 'rxjs';
import { CookieService } from '../cookie.service';
import { CookieInstances } from '../../objects/cookie-instances';

export const authGuard: CanActivateFn = (
  route,
  snapshot
): Observable<boolean | UrlTree> | Promise<boolean | UrlTree> | boolean | UrlTree => {
  const cookieService = inject(CookieService);
  const router = inject(Router);

  const jwt = cookieService.get(CookieInstances.Jwt);

  if (!jwt) {
    cookieService.delete(CookieInstances.Jwt);
    return router.navigate(['/', 'login']);
  }

  // const jwtParts = jwt.split('.');
  // if (jwtParts.length !== 3) {
  //   cookieService.delete(CookieInstances.Jwt);
  //   return router.navigate(['/', 'login']);
  // }
  //
  // const payloadBase64 = jwtParts[1];
  //
  // try {
  //   const payloadJson = atob(payloadBase64);
  //   const payload = JSON.parse(payloadJson);
  //
  //   if (!payload.exp || payload.exp * 1000 < Date.now()) {
  //     cookieService.delete(CookieInstances.Jwt);
  //     return router.navigate(['/', 'login']);
  //   }
  //
  //   return true;
  // } catch {
  //   cookieService.delete(CookieInstances.Jwt);
  //   return router.navigate(['/', 'login']);
  // }

  return true;
};
