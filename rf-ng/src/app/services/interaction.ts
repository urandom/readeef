import { Injectable, Output, EventEmitter } from '@angular/core'
import { Location } from '@angular/common';
import { Router } from '@angular/router';

@Injectable({providedIn: "root"})
export class InteractionService {
    @Output() toolbarTitleClickEvent = new EventEmitter<void>();

    constructor(
        private location: Location,
        private router: Router,
    ) {}

    toolbarTitleClick() {
        this.toolbarTitleClickEvent.emit();
    }

    navigateUp(): boolean {
        let path = this.location.path()
        let idx = path.indexOf("/article/")
        if (idx != -1) {
            this.router.navigateByUrl(path.substring(0, idx), {state: {"articleID": path.substring(idx+9)}});
            return true;
        }
        return false;
    }
}
