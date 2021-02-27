import { Injectable } from '@angular/core'
import { Observable ,  BehaviorSubject } from "rxjs";


interface Prefs  {
    olderFirst: boolean;
    unreadOnly: boolean;
    unreadFirst: boolean;
    searchOrder: SearchOrder;
}

export interface ListPreferences {
    olderFirst: boolean
    unreadOnly: boolean
    unreadFirst: boolean
    searchOrder: SearchOrder
}

type SearchOrder = "default" | "olderFirst" | "newerFirst";

@Injectable({providedIn: "root"})
export class PreferencesService {
    private prefs : Prefs;
    private queryPreferencesSubject : BehaviorSubject<ListPreferences>;
    private static key = "preferences";

    constructor() {
        this.prefs = 
            JSON.parse(localStorage.getItem(PreferencesService.key) || "{}");

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

    get searchOrder(): SearchOrder {
        if (!this.prefs.searchOrder) {
            return this.prefs.olderFirst ? "olderFirst" : "newerFirst";
        }
        return this.prefs.searchOrder;
    }

    set searchOrder(val: SearchOrder) {
        this.prefs.searchOrder = val;
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
