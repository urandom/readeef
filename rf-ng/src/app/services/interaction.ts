import { Injectable, Output, EventEmitter } from '@angular/core'

@Injectable({providedIn: "root"})
export class InteractionService {
    @Output() toolbarTitleClickEvent = new EventEmitter<void>();

    constructor() {}

    toolbarTitleClick() {
        this.toolbarTitleClickEvent.emit();
    }
}
