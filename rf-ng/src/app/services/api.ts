import { Injectable } from '@angular/core'
import { Location } from '@angular/common'
import { HttpClient, HttpResponse, HttpErrorResponse, HttpHeaders } from '@angular/common/http';
import { Router } from '@angular/router';
import { environment } from "../../environments/environment"
import { Observable, pipe ,  range ,  timer ,  of, empty} from "rxjs";
import { catchError, flatMap, retryWhen, map, zip } from "rxjs/operators"

@Injectable({
    providedIn: "root",
})
export class APIService {
    constructor(
        private http: HttpClient,
        private router: Router,
    ) { }

    get<T>(endpoint: string, headers?: HttpHeaders) : Observable<T> {
        return this.http.get<T>(
            absEndpoint(endpoint),
            { headers: authHeaders(headers) }
        ).pipe(
            catchError(unathorizeHandler(this.router)),
            retryWhen(errorRetry())
        );
    }

    post<T>(endpoint: string, body?: any, headers?: HttpHeaders) : Observable<T>  {
        return this.http.post<T>(
            absEndpoint(endpoint),
            body,
            { headers: authHeaders(headers) }
        ).pipe(
            catchError(unathorizeHandler(this.router)),
            retryWhen(errorRetry())
        );
    }

    rawPost(endpoint: string, body?: any, headers?: HttpHeaders) : Observable<HttpResponse<string>>  {
        return this.http.post(
            absEndpoint(endpoint),
            body,
            { headers: authHeaders(headers), observe: "response", responseType: "text" }
        ).pipe(
            catchError(unathorizeHandler(this.router)),
            retryWhen(errorRetry())
        );
    }

    delete<T>(endpoint: string, headers?: HttpHeaders) : Observable<T> {
        return this.http.delete<T>(
            absEndpoint(endpoint),
            { headers: authHeaders(headers) }
        ).pipe(
            catchError(unathorizeHandler(this.router)),
            retryWhen(errorRetry())
        );
    }

    put<T>(endpoint: string, body: any, headers?: HttpHeaders) : Observable<T> {
        return this.http.put<T>(
            absEndpoint(endpoint),
            body,
            { headers: authHeaders(headers) }
        ).pipe(
            catchError(unathorizeHandler(this.router)),
            retryWhen(errorRetry())
        );
    }
}

export function unathorizeHandler<T>(router: Router) : (err: HttpErrorResponse, caught: Observable<T>) => Observable<T> {
    return (err: HttpErrorResponse, caught: Observable<T>) => {
        return of(err).pipe(flatMap(response => {
            if (response.status == 403 || response.status == 401) {
                if (!router.routerState.snapshot.url.startsWith("/login")) {
                    router.navigate(['/login'], { queryParams: { returnUrl: router.url } });
                }
                return empty();
            }


            throw response;
        }))
    }
}

function errorRetry() : (errors: Observable<HttpErrorResponse>) => Observable<any> {
    return errors => range(1, 10).pipe(
        zip(errors, (i, err) => {
            if (i == 10) {
                throw err;
            }

            return i;
        }),
        flatMap(i => timer(i * 1000)),
    )
}

let absEndpoint = function (endpoint: string): string {
    return environment.apiEndpoint + endpoint;
}

let authHeaders = function(headers?: HttpHeaders) : HttpHeaders {
    headers = headers ? headers : new HttpHeaders();

    return headers.set(
        "Authorization", localStorage.getItem("token")
    ).set(
        "Accept", "application/json"
    );
}
