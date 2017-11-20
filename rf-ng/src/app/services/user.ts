import { Injectable } from "@angular/core"
import { Headers } from '@angular/http'
import { APIService, Serializable } from "./api";
import { Observable } from "rxjs";
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/shareReplay'

export class User {
    login: string
    firstName: string
    lastName: string
    email: string
    admin: boolean
    active: boolean
    profileData: Map<string, any>
}

export interface PasswordChange {
    current: string
    new: string
}

class UserResponse extends Serializable {
    user: User

    fromJSON(json) {
        if ("profileData" in json) {
            let data = json["profileData"];
            json["profileData"] = new Map<string, any>(data);
        }

        return super.fromJSON(json);
    }
}

class UsersResponse extends Serializable {
    users: User[]
}

class SettingsResponse extends Serializable {
    success: boolean
}

export interface AddUser {
    login: string
    password: string
}

@Injectable()
export class UserService {
    user : Observable<User>

    constructor(private api: APIService) {
        this.user = this.api.get("user/current").map(response =>
            new UserResponse().fromJSON(response.json()).user
        ).shareReplay(1)
    }

    getCurrentUser() : Observable<User> {
        return this.user
    }

    changeUserPassword(updated: string, current: string) : Observable<boolean> {
        return this.setUserSetting("password", updated, {"current": current});
    }

    setUserSetting(key: string, value: any, extra?: {[key: string]: string}) : Observable<boolean> {
        var data = `value=${encodeURIComponent(value)}`;
        if (extra) {
            for (let key in extra) {
                data += `&${key}=${encodeURIComponent(extra[key])}`;
            }
        }
        return this.api.put(`user/settings/${key}`, data, new Headers({
            "Content-Type": "application/x-www-form-urlencoded",
        })).map(
            response => {
                if (response.ok) {
                    return new SettingsResponse().fromJSON(response.json()).success
                } else if (response.status >= 500) {
                    throw new Error("Error: " + response.text())
                }

                return false
            }
        )
    }

    list() : Observable<User[]> {
        return this.api.get("user").map(response =>
            new UsersResponse().fromJSON(response.json()).users
        );
    }

    addUser(login: string, password: string) : Observable<boolean> {
        var body = new FormData();
        body.append("login", login);
        body.append("password", password);

        return this.api.post("user", body).map(
            response => !!response.json()["success"]
        );
    }

    deleteUser(user: string) : Observable<boolean> {
        return this.api.delete(`user/${user}`).map(response =>
            !!response.json()["success"]
        );
    }

    toggleActive(user: string, active: boolean) : Observable<boolean> {
        var data = `value=${active}`
        return this.api.put(
            `user/${user}/settings/is-active`, data, new Headers({
                "Content-Type": "application/x-www-form-urlencoded",
            })
        ).map(response => !!response.json()["success"]);
    }
}