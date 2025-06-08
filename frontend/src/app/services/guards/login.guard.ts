import { CanActivateFn, Router, UrlTree } from '@angular/router';
import { inject } from '@angular/core';
import { Observable } from 'rxjs';
import { CookieInstances } from '../../objects/cookie-instances';
import { CookieService } from '../cookie.service';

export const loginGuard: CanActivateFn = (
  route,
  snapshot
): Observable<boolean | UrlTree> | Promise<boolean | UrlTree> | boolean | UrlTree => {
  const cookieService = inject(CookieService);
  const router = inject(Router);

  if (cookieService.get(CookieInstances.Jwt)) {
    return router.navigate(['/']);
  }

  return true;
};
