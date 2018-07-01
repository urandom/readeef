import { Router, NavigationEnd, ActivatedRouteSnapshot } from '@angular/router';
import { Observable, pipe } from "rxjs";
import { filter, map, startWith, distinctUntilChanged, shareReplay } from 'rxjs/operators';

export function listRoute(router: Router) : Observable<ActivatedRouteSnapshot> {
    return router.events.pipe(
        filter(event => event instanceof NavigationEnd),
        map(v => {
            return getListRoute([router.routerState.snapshot.root])
        }),
        startWith(getListRoute([router.routerState.snapshot.root])),
        distinctUntilChanged(),
        shareReplay(1),
    );
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

export function articleRoute(router: Router): Observable<ActivatedRouteSnapshot> {
    return router.events.pipe(
        filter(event => event instanceof NavigationEnd),
        map(v => {
            return getArticleRoute([router.routerState.snapshot.root])
        }),
        startWith(getArticleRoute([router.routerState.snapshot.root])),
        distinctUntilChanged(),
        shareReplay(1),
    );
}

export function getArticleRoute(routes: ActivatedRouteSnapshot[]): ActivatedRouteSnapshot {
    for (let route of routes) {
        if ("articleID" in route.params) {
            return route;
        }

        let r = getArticleRoute(route.children);
        if (r != null) {
            return r;
        }
    }

    return null;
}