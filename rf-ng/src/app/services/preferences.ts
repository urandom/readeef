import { Injectable } from '@angular/core'
import { Serializable } from "./api"

class Prefs extends Serializable {
    olderFirst: boolean
    unreadOnly: boolean
}

@Injectable()
export class PreferencesService {
    private prefs = new Prefs()

    private static key = "preferences"

    constructor() {
        this.prefs.fromJSON(
            JSON.parse(localStorage.getItemt(PreferencesService.key))
        )
    }

    get olderFirst() : boolean {
        return this.prefs.olderFirst;
    }

    set olderFirst(val: boolean) {
        this.prefs.olderFirst = val;

        this.saveToStorage();
    }

    get unreadOnly() : boolean {
        return this.prefs.unreadOnly;
    }

    set unreadOnly(val: boolean) {
        this.prefs.unreadOnly = val;

        this.saveToStorage();
    }

    private saveToStorage() {
        localStorage.setItem(PreferencesService.key, JSON.stringify(this.prefs))
    }
}