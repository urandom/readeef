import { Injectable } from '@angular/core'
import { Http, Headers, Response } from '@angular/http'
import { environment } from "../../environments/environment"
import { Observable, BehaviorSubject } from "rxjs";
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/skip'

@Injectable()
export class TokenService {
    private tokenSubject : BehaviorSubject<string>

    constructor(private http: Http) {
        this.tokenSubject = new BehaviorSubject(localStorage.getItem("token"));

        this.tokenSubject.skip(1).subscribe(token => {
            if (token == "") {
                localStorage.removeItem("token");
            } else {
                localStorage.setItem("token", token);
            }
        })
    }

    create(user: string, password: string): Observable<string> {
        var body = new FormData();
        body.append("user", user);
        body.append("password", password);

        return this.http.post(environment.apiEndpoint + "token", body)
            .map(response => {
                return response.headers.get("Authorization");
            })
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
        )
    }
}

export class AuthenticationError extends Error {
}