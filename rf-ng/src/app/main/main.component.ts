import { Component, OnInit, OnDestroy, ViewChild } from "@angular/core";
import { NavigationEnd, Router, RouterEvent } from '@angular/router';
import { Subscription } from "rxjs";
import { listRoute, articleRoute } from "./routing-util"
import { TokenService } from "../services/auth";
import { filter, switchMap, map, combineLatest } from "rxjs/operators";
import { MatSidenav } from '@angular/material/sidenav';

@Component({
    templateUrl: "./main.html",
    styleUrls: ["./main.css"],
})
export class MainComponent implements OnInit, OnDestroy {
    showsArticle = false
    inSearch = false

    private subscriptions = new Array<Subscription>();

    @ViewChild('sidenav', {static: true})
    private sidenav: MatSidenav;

    constructor(
        private tokenService: TokenService,
        private router: Router,
    ) {}

    ngOnInit() {
        this.subscriptions.push(this.tokenService.tokenObservable().pipe(
            filter(token => token != ""),
            switchMap(_ =>
                articleRoute(this.router).pipe(
                    map(route => route != null),
                    combineLatest(
                        listRoute(this.router).pipe(map(route =>
                            route != null && route.data["primary"] == "search"
                        )),
                        (inArticles, inSearch) : [boolean, boolean] =>
                            [inArticles, inSearch]
                    ),
                ),
            ),
        ).subscribe(
            data => {
                this.showsArticle = data[0];
                this.inSearch = data[1];
            },
            error => console.log(error)
        ));

        this.subscriptions.push(this.router.events.pipe(
           filter(e => e instanceof NavigationEnd)
        ).subscribe(_ => {
            this.sidenav.close();
        }, err => console.log(err)));
    }

    ngOnDestroy(): void {
        this.subscriptions.forEach(s => s.unsubscribe());
    }
}
