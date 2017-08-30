import { Component, OnInit } from '@angular/core';
import { PreferencesService } from "../services/preferences";

@Component({
  templateUrl: './toolbar-feed.html',
  styleUrls: ['./toolbar.css']
})
export class ToolbarFeedComponent implements OnInit {
    olderFirst = false

    constructor(private preferences : PreferencesService) {}

    ngOnInit(): void {
    }

    toggleOlderFirst() {
        this.preferences.olderFirst = !this.preferences.olderFirst;
        this.olderFirst = this.preferences.olderFirst;
    }

    toggleUnreadOnly() {
        this.preferences.unreadOnly = !this.preferences.unreadOnly;
    }
}