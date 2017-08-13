import { Injectable } from '@angular/core';
import { Router, NavigationStart } from '@angular/router';
import { Subject } from "rxjs/Subject";

@Injectable()
export class AlertService {
    private subject = new Subject<any>();
    private keepAfterNavigationChange = false;

    constructor(private router: Router) {
        router.events
            .subscribe(event => {
                if (event instanceof NavigationStart) {
                    if (this.keepAfterNavigationChange) {
                        this.keepAfterNavigationChange = false;
                    } else {
                        this.subject.next();
                    }
                }
            })
    }

    message(message: string, success = true, keepAfterNavigationChange = false) {
        this.keepAfterNavigationChange = keepAfterNavigationChange;
        this.subject.next({ type: success ? 'success' : 'error', text: message });
    }

    getMessage() {
        return this.subject.asObservable();
    }
}