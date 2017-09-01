import { Injectable } from '@angular/core'
import { Http, Headers, Response } from '@angular/http'
import { Router } from '@angular/router';
import { environment } from "../../environments/environment"
import { Observable } from "rxjs/Observable";
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/mergeMap';
import 'rxjs/add/operator/catch';
import 'rxjs/add/observable/empty';
import 'rxjs/add/observable/of';

export class Serializable {
    fromJSON(json) {
        for (var propName in json)
            this[propName] = json[propName];
        return this;
    }
}

@Injectable()
export class APIService {
    constructor(private http: Http, private router: Router) { }

    get(endpoint: string, headers?: Headers) {
        return this.http.get(
            absEndpoint(endpoint),
            { headers: authHeaders(headers) }
        )
            .map(checkRequestForToken)
            .catch(unathorizeHandler(this.router));
    }

    post(endpoint: string, body?: any, headers?: Headers) {
        return this.http.post(
            absEndpoint(endpoint),
            body,
            { headers: authHeaders(headers) }
        )
            .map(checkRequestForToken)
            .catch(unathorizeHandler(this.router));
    }

    delete(endpoint: string, headers?: Headers) {
        return this.http.delete(
            absEndpoint(endpoint),
            { headers: authHeaders(headers) }
        )
            .map(checkRequestForToken)
            .catch(unathorizeHandler(this.router));
    }

    put(endpoint: string, body: any, headers?: Headers) {
        return this.http.put(
            absEndpoint(endpoint),
            body,
            { headers: authHeaders(headers) }
        )
            .map(checkRequestForToken)
            .catch(unathorizeHandler(this.router));
    }
}

export function unathorizeHandler(router: Router) : (response: Response) => Observable<Response> {
    return (response: Response) => {
        return Observable.of(response).flatMap(response => {
            if (response.status == 401) {
                router.navigate(['/login'], { queryParams: { returnUrl: router.url } });
                return Observable.empty<Response>();
            }

            return Observable.of(response);
        })
    }
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

let checkRequestForToken = function(response: Response) : Response {
    var token = response.headers.get("Authorization");
    if (token) {
        localStorage.setItem("token", token);
    }

    return response;
}