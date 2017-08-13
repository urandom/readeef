import { Component, OnInit } from '@angular/core'
import { AlertService } from "../services/alert"

@Component({
    moduleId: module.id,
    selector: "alert",
    templateUrl: "./alert.html",
})
export class AlertComponent implements OnInit {
    message: any;

    constructor(private alertService: AlertService) {}

    ngOnInit() {
        this.alertService.getMessage().subscribe(message => this.message = message);
    }
}