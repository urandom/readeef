import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from "../../environments/environment";
import { BehaviorSubject, Observable, timer, never, interval } from "rxjs";


import { APIService } from './api';
import { skip, switchMap, map, filter, catchError, throttle } from 'rxjs/operators';

@Injectable({
    providedIn: "root",
})
export class TokenService {
    private tokenSubject : BehaviorSubject<string>

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

		// Renew the token once a day
		this.tokenSubject.pipe(
            switchMap(t => t == "" ? never() : timer(0, 3600000)),
            map(v => new Date(localStorage.getItem("token_age"))),
            filter(date => new Date().getTime() - date.getTime() > 86400000),
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

        return this.http.post(environment.apiEndpoint + "token", body, {observe: "response", responseType: "text"}).pipe(
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

    tokenObservable() : Observable<string> {
        return this.tokenSubject.pipe(
            map(auth => (auth || "").replace("Bearer ", "")),
            throttle(v => interval(2000)),
        );
    }
}

export class AuthenticationError extends Error {
}
