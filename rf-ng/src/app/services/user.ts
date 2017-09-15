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
}

class UserResponse extends Serializable {
    user: User
}

class SettingsResponse extends Serializable {
    success: boolean
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
}