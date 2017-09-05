import { Component, OnInit, OnDestroy } from "@angular/core";
import { Router } from '@angular/router';
import { Subscription } from "rxjs";
import { articleRoute } from "./routing-util"

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
    ) {}

    ngOnInit() {
        this.subscription = articleRoute(
            this.router
        ).map(route => route != null).subscribe(
            showsArticle => this.showsArticle = showsArticle
        );
    }

    ngOnDestroy(): void {
        this.subscription.unsubscribe();
    }
}