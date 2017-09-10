import { Component, OnInit, OnDestroy } from "@angular/core";
import { Router } from '@angular/router';
import { Subscription } from "rxjs";
import { listRoute, articleRoute } from "./routing-util"

@Component({
    moduleId: module.id,
    templateUrl: "./main.html",
    styleUrls: ["./main.css"],
})
export class MainComponent implements OnInit, OnDestroy {
    showsArticle = false
    inSearch = false

    private subscriptions = new Array<Subscription>()

    constructor(
        private router: Router,
    ) {}

    ngOnInit() {
        this.subscriptions.push(articleRoute(
            this.router
        ).map(route => route != null).subscribe(
            showsArticle => this.showsArticle = showsArticle
        ))

        this.subscriptions.push(listRoute(this.router).map(
            route => route != null && route.data["primary"] == "search"
        ).subscribe(
            inSearch => this.inSearch = inSearch,
            error => console.log(error),
        ))
    }

    ngOnDestroy(): void {
        for (let subscription of this.subscriptions) {
            subscription.unsubscribe()
        }
    }
}