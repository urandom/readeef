import { Component, OnInit, OnDestroy } from "@angular/core";
import { Location } from '@angular/common';
import { Router } from '@angular/router';
import { Subscription } from "rxjs";
import { articleDisplayRoute } from "./routing-util"

@Component({
    moduleId: module.id,
    templateUrl: "./main.html",
    styleUrls: ["./main.css"],
})
export class MainComponent implements OnInit, OnDestroy {
    showsArticle = false

    private subscription : Subscription

    constructor(
        private router: Router,
        private location: Location,
    ) {}

    ngOnInit() {
        this.subscription = articleDisplayRoute(
            this.router, this.location
        ).subscribe(
            showsArticle => this.showsArticle = showsArticle
        );
    }

    ngOnDestroy(): void {
        this.subscription.unsubscribe();
    }
}