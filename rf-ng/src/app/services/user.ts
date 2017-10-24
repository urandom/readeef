import { Injectable } from "@angular/core"
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

    changeUserPassword(value: PasswordChange) : Observable<boolean> {
        return this.setUserSetting("password", value)
    }

    setUserSetting(key: string, value: any) : Observable<boolean> {
        return this.api.put(`user/settings/${key}`, JSON.stringify(value)).map(
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

    addUser(data: AddUser) : Observable<boolean> {
        return this.api.post("user", JSON.stringify(data)).map(
            response => !!response.json()["success"]
        );
    }

    deleteUser(user: string) : Observable<boolean> {
        return this.api.delete(`user/${user}`).map(response =>
            !!response.json()["success"]
        );
    }

    toggleActive(user: string, active: boolean) : Observable<boolean> {
        return this.api.put(
            `user/${user}/settings/is-active`, JSON.stringify(active)
        ).map(response => !!response.json()["success"]);
    }

    setProfile(user: string, profile: Map<string, any>) : Observable<boolean> {
        return this.api.put(
            `user/${user}/settings/profile`, JSON.stringify(profile)
        ).map(response => !!response.json()["success"]);
    }
}