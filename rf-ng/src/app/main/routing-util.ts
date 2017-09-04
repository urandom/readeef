import { Router, NavigationEnd, ActivatedRouteSnapshot } from '@angular/router';
import { Location } from '@angular/common';
import { Observable } from "rxjs";
import 'rxjs/add/operator/distinctUntilChanged'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/startWith'
import 'rxjs/add/operator/shareReplay'

export const articlePattern = "/article/";

export function articleDisplayRoute(router: Router, location: Location): Observable<boolean> {
    return router.events.filter(event =>
        event instanceof NavigationEnd
    ).startWith(null).map(v => {
        return location.path().indexOf(articlePattern) != -1
    }).distinctUntilChanged().shareReplay(1);
}

export function getListRoute(routes: ActivatedRouteSnapshot[]): ActivatedRouteSnapshot {
    for (let route of routes) {
        if ("primary" in route.data) {
            return route;
        }

        let r = getListRoute(route.children);
        if (r != null) {
            return r;
        }
    }

    return null;

}