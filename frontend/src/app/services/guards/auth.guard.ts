import { CanActivateFn, Router, UrlTree } from '@angular/router';
import { inject } from '@angular/core';
import { Observable } from 'rxjs';
import { CookieService } from '../cookie.service';
import { CookieInstances } from '../../objects/cookie-instances';

export const authGuard: CanActivateFn = (): Observable<boolean | UrlTree> | Promise<boolean | UrlTree> | boolean | UrlTree => {
  const cookieService = inject(CookieService);
  const router = inject(Router);

  const jwt = cookieService.get(CookieInstances.Jwt);

  if (!jwt) {
    cookieService.delete(CookieInstances.Jwt);
    return router.navigate(['/', 'login']);
  }

  return true;
};
