import { Component, OnInit } from "@angular/core" ;
import { MatSlideToggle } from "@angular/material";
import { SharingService, ShareService } from "../../services/sharing"

@Component({
    selector: "settings-share-services",
    templateUrl: "./share-services.html",
    styleUrls: ["../common.css", "./share-services.css"]
})
export class ShareServicesSettingsComponent implements OnInit {
    services: [ShareService, boolean][][]

    constructor(
        private sharingService: SharingService,
    ) {}

    ngOnInit(): void {
        this.services = this.sharingService.groupedList();
    }

    toggleService(id: string, enabled: boolean) {
        this.sharingService.toggle(id, enabled);
    }
}