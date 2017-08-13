import { Injectable } from '@angular/core'
import { Http, Headers, Response } from '@angular/http'
import { environment } from "../../environments/environment"
import { Observable } from "rxjs";

@Injectable()
export class TokenService {
    constructor(private http: Http) { }

    create(user: string, password: string): Observable<string> {
        return this.http.post(
            environment.apiEndpoint + "token",
             "user=${user}&password=${password}"
        ).map(
            response => response.headers.get("Authorization")
        ).map(
            auth => {
                localStorage.setItem("token", auth);

                return auth;
            }
        );
    }

    delete() {
        localStorage.removeItem("token");
    }
}