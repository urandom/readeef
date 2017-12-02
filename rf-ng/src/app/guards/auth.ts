import { Injectable } from '@angular/core'
import { Router, CanActivate, ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router'
import * as jwt from 'jwt-decode'

@Injectable()
export class AuthGuard implements CanActivate {
    constructor(private router: Router) {}

    canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot) : boolean {
        var token = localStorage.getItem("token")
        if (token) {
            var res = jwt(token)
            if (
                (!res["exp"] || res["exp"] >= new Date().getTime() / 1000) &&
                (!res["nbf"] || res["nbf"] < new Date().getTime() / 1000)
            )  {
                return true;
            }
        }

        this.router.navigate(['/login'], { queryParams: { returnUrl: state.url }});
        return false;
    }
}