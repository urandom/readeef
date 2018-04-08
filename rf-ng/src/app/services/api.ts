import { Injectable } from '@angular/core'
import { Location } from '@angular/common'
import { Http, Headers, Response, ResponseType } from '@angular/http'
import { Router } from '@angular/router';
import { environment } from "../../environments/environment"
import { Observable } from "rxjs/Observable";
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/mergeMap';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/retryWhen';
import 'rxjs/add/operator/zip';
import 'rxjs/add/observable/empty';
import 'rxjs/add/observable/range';
import 'rxjs/add/observable/of';
import 'rxjs/add/observable/timer';

export class Serializable {
    fromJSON(json) {
        for (var propName in json)
            this[propName] = json[propName];
        return this;
    }
}

@Injectable()
export class APIService {
    constructor(
        private http: Http,
        private router: Router,
    ) { }

    get(endpoint: string, headers?: Headers) {
        return this.http.get(
            absEndpoint(endpoint),
            { headers: authHeaders(headers) }
        )
            .catch(unathorizeHandler(this.router))
            .retryWhen(errorRetry());
    }

    post(endpoint: string, body?: any, headers?: Headers) {
        return this.http.post(
            absEndpoint(endpoint),
            body,
            { headers: authHeaders(headers) }
        )
            .catch(unathorizeHandler(this.router))
            .retryWhen(errorRetry());
    }

    delete(endpoint: string, headers?: Headers) {
        return this.http.delete(
            absEndpoint(endpoint),
            { headers: authHeaders(headers) }
        )
            .catch(unathorizeHandler(this.router))
            .retryWhen(errorRetry());
    }

    put(endpoint: string, body: any, headers?: Headers) {
        return this.http.put(
            absEndpoint(endpoint),
            body,
            { headers: authHeaders(headers) }
        )
            .catch(unathorizeHandler(this.router))
            .retryWhen(errorRetry());
    }
}

export function unathorizeHandler(router: Router) : (response: Response) => Observable<Response> {
    return (response: Response) => {
        return Observable.of(response).flatMap(response => {
            if (response.status == 403 || response.status == 401) {
                if (!router.routerState.snapshot.url.startsWith("/login")) {
                    router.navigate(['/login'], { queryParams: { returnUrl: router.url } });
                }
                return Observable.empty<Response>();
            }

            if (response.type == ResponseType.Error) {
                throw response;
            }

            return Observable.of(response);
        })
    }
}

function errorRetry() : (errors: Observable<Response>) => Observable<any> {
    return errors => Observable.range(1, 10).zip(errors, (i, err) => {
        if (i == 10) {
            throw err;
        }

        return i;
    }).flatMap(i => Observable.timer(i * 1000));
}

let absEndpoint = function (endpoint: string): string {
    return environment.apiEndpoint + endpoint;
}

let authHeaders = function(headers?: Headers) : Headers {
    var headers = new Headers(headers);
    headers.set("Authorization", localStorage.getItem("token"));
    headers.set("Accept", "application/json");

    return headers;
}
