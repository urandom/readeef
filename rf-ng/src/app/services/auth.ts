import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, interval, never, Observable, timer } from "rxjs";
import { catchError, filter, map, switchMap, throttle } from 'rxjs/operators';
import { environment } from "../../environments/environment";
import { APIService } from './api';



@Injectable({
    providedIn: "root",
})
export class TokenService {
    private tokenSubject: BehaviorSubject<string>

    constructor(
        private http: HttpClient,
        private api: APIService,
    ) {
        this.tokenSubject = new BehaviorSubject(localStorage.getItem("token"));

        this.tokenSubject.pipe(
            filter(token => token != localStorage.getItem("token"))
        ).subscribe(token => {
            if (token == "") {
                localStorage.removeItem("token");
            } else {
                localStorage.setItem("token", token);
                localStorage.setItem("token_age", new Date().toString())
            }
        });

        if (!localStorage.getItem("token_age") && localStorage.getItem("token")) {
            localStorage.setItem("token_age", new Date().toString());
        }

        // Renew the token once every day
        this.tokenSubject.pipe(
            switchMap(t => t == "" ? never() : timer(0, 3600000)),
            map(v => new Date(localStorage.getItem("token_age"))),
            filter(date => new Date().getTime() - date.getTime() > 144000000),
            switchMap(
                v => {
                    console.log("Renewing user token");
                    return this.api.rawPost("user/token").pipe(
                        map(r => r.headers.get("Authorization")),
                        filter(token => token != ""),
                        catchError(err => {
                            console.log("Error generating new user token: ", err);
                            return never();
                        }),
                    );
                }
            ),
        ).subscribe(token => this.tokenSubject.next(token));
    }

    create(user: string, password: string): Observable<string> {
        var body = new FormData();
        body.append("user", user);
        body.append("password", password);

        return this.http.post(environment.apiEndpoint + "token", body, { observe: "response", responseType: "text" }).pipe(
            map(response => {
                return response.headers.get("Authorization");
            }),
            map(auth => {
                if (auth) {
                    this.tokenSubject.next(auth);

                    return auth;
                }

                throw new AuthenticationError("test");
            }),
        );
    }

    delete() {
        this.tokenSubject.next("");
    }

    tokenObservable(): Observable<string> {
        return this.tokenSubject.pipe(
            map(auth => (auth || "").replace("Bearer ", "")),
            throttle(v => interval(2000)),
        );
    }

    tokenUser(token: string): string {
        if (!token) {
            return ""
        }

        var payload = JSON.parse(atob(token.split(".")[1]));
        return payload["sub"];
    }
}

export class AuthenticationError extends Error {
}
