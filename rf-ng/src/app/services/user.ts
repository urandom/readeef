import { Injectable } from "@angular/core"
import { APIService } from "./api";
import { Observable } from "rxjs";


import { HttpHeaders } from "@angular/common/http";
import { map, shareReplay } from "rxjs/operators";

export class User {
    login: string;
    firstName: string;
    lastName: string;
    email: string;
    admin: boolean;
    active: boolean;
    profileData?: Map<string, any>;
}

export interface PasswordChange {
    current: string;
    new: string;
}

interface UserResponse {
    user: User;
}

interface UsersResponse {
    users: User[];
}

interface SettingsResponse  {
    success: boolean;
}

export interface AddUser {
    login: string
    password: string
}

@Injectable({providedIn: "root"})
export class UserService {
    user : Observable<User>

    constructor(private api: APIService) {
        this.user = this.api.get<UserResponse>("user/current").pipe(
            map(response => response.user),
            shareReplay(1)
        );
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
        return this.api.put<SettingsResponse>(`user/settings/${key}`, data, new HttpHeaders({
            "Content-Type": "application/x-www-form-urlencoded",
        })).pipe(
            map( response => response.success)
        );
    }

    list() : Observable<User[]> {
        return this.api.get<UsersResponse>("user").pipe(
            map(response => response.users)
        );
    }

    addUser(login: string, password: string) : Observable<boolean> {
        var body = new FormData();
        body.append("login", login);
        body.append("password", password);

        return this.api.post<SettingsResponse>("user", body).pipe(
            map(response => response.success)
        );
    }

    deleteUser(user: string) : Observable<boolean> {
        return this.api.delete<SettingsResponse>(`user/${user}`).pipe(
            map(response => response.success)
        );
    }

    toggleActive(user: string, active: boolean) : Observable<boolean> {
        var data = `value=${active}`
        return this.api.put<SettingsResponse>(
            `user/${user}/settings/is-active`, data, new HttpHeaders({
                "Content-Type": "application/x-www-form-urlencoded",
            })
        ).pipe(
            map(response => response.success)
        );
    }
}