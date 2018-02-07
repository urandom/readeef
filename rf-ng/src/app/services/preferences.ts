import { Injectable } from '@angular/core'
import { Serializable } from "./api"
import { Observable } from "rxjs";
import { BehaviorSubject } from "rxjs/BehaviorSubject";
import 'rxjs/add/operator/distinctUntilChanged'

class Prefs extends Serializable {
    olderFirst: boolean
    unreadOnly: boolean
    unreadFirst: boolean
}

export interface ListPreferences {
    olderFirst: boolean
    unreadOnly: boolean
    unreadFirst: boolean
}

@Injectable()
export class PreferencesService {
    private prefs = new Prefs()
    private queryPreferencesSubject : BehaviorSubject<ListPreferences>
    private static key = "preferences"

    constructor() {
        this.prefs.fromJSON(
            JSON.parse(localStorage.getItem(PreferencesService.key))
        )

        this.prefs.unreadFirst = true;

        this.queryPreferencesSubject = new BehaviorSubject(this.prefs);
    }

    get olderFirst() : boolean {
        return this.prefs.olderFirst;
    }

    set olderFirst(val: boolean) {
        this.prefs.olderFirst = val;
        this.queryPreferencesSubject.next(this.prefs);

        this.saveToStorage();
    }

    get unreadOnly() : boolean {
        return this.prefs.unreadOnly;
    }

    set unreadOnly(val: boolean) {
        this.prefs.unreadOnly = val;
        this.queryPreferencesSubject.next(this.prefs);

        this.saveToStorage();
    }

    queryPreferences() : Observable<ListPreferences> {
        return this.queryPreferencesSubject.asObservable();
    }

    private saveToStorage() {
        localStorage.setItem(PreferencesService.key, JSON.stringify(this.prefs))
    }
}