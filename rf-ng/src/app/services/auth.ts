import { Injectable } from '@angular/core'
import { Http, Headers, Response } from '@angular/http'
import { environment } from "../../environments/environment"
import { Observable } from "rxjs";
import 'rxjs/add/operator/map'

@Injectable()
export class TokenService {
    constructor(private http: Http) { }

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
                    localStorage.setItem("token", auth);

                    return auth;
                }

                throw new AuthenticationError("test");
            });
    }

    delete() {
        localStorage.removeItem("token");
    }
}

export class AuthenticationError extends Error {
}