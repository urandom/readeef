import { Injectable } from '@angular/core'
import { Http, Headers, Response } from '@angular/http'
import { environment } from "../../environments/environment"
import { BehaviorSubject, Observable } from "rxjs";
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/skip'
import { APIService } from './api';

@Injectable()
export class TokenService {
    private tokenSubject : BehaviorSubject<string>

    constructor(
		private http: Http,
		private api: APIService,
	) {
        this.tokenSubject = new BehaviorSubject(localStorage.getItem("token"));

        this.tokenSubject.skip(1).subscribe(token => {
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
		this.tokenSubject.switchMap(
			t => t == "" ? Observable.never() : Observable.timer(0, 3600000)
		).map(
			v => new Date(localStorage.getItem("token_age"))
		).filter(
			date => new Date().getTime() - date.getTime() > 86400000
		).switchMap(
			v => this.api.post("user/token")
				.map(r => r.headers.get("Authorization"))
				.filter(token => token != "")
				.catch(err => {
					console.log("Error generating new user token: ", err);
					return Observable.never<string>();
				})
		).subscribe(token => this.tokenSubject.next(token));
    }

    create(user: string, password: string): Observable<string> {
        var body = new FormData();
        body.append("user", user);
        body.append("password", password);

        return this.http.post(environment.apiEndpoint + "token", body)
            .map(
				response => response.headers.get("Authorization")
            )
            .map(auth => {
                if (auth) {
                    this.tokenSubject.next(auth);

                    return auth;
                }

                throw new AuthenticationError("test");
            });
    }

    delete() {
        this.tokenSubject.next("");
    }

    tokenObservable() : Observable<string> {
        return this.tokenSubject.map(auth =>
             (auth || "").replace("Bearer ", "")
        ).throttle(v => Observable.interval(2000));
    }
}

export class AuthenticationError extends Error {
}
