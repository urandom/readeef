import { Component, OnInit, OnDestroy, ViewChild } from '@angular/core';
import { Location } from '@angular/common';
import { Article, ArticleService } from "../services/article";
import { FeaturesService } from "../services/features";
import { PreferencesService } from "../services/preferences";
import { SharingService, ShareService } from "../services/sharing";
import { Router, NavigationStart } from '@angular/router';
import { Observable, Subscription } from "rxjs";
import { articleRoute, listRoute, getListRoute } from "../main/routing-util"
import 'rxjs/add/observable/empty'
import 'rxjs/add/observable/of'
import 'rxjs/add/observable/timer'
import 'rxjs/add/operator/delayWhen'
import 'rxjs/add/operator/distinctUntilChanged'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/shareReplay'
import 'rxjs/add/operator/startWith'
import 'rxjs/add/operator/switchMap'
import 'rxjs/add/operator/take'

@Component({
  templateUrl: './toolbar-feed.html',
  styleUrls: ['./toolbar.css'],
  host: {
      '(keydown.enter)': 'keyEnter()',
  }
})
export class ToolbarFeedComponent implements OnInit, OnDestroy {
    olderFirst = false
    showsArticle = false
    articleRead = false
    searchButton = false
    markAllRead = false
    inSearch = false
    enabledShares = false

    shareServices = new Array<ShareService>()

    private _searchQuery = ""
    private _searchEntry = false

    private articleID : Observable<number>
    private subscriptions = new Array<Subscription>()

    @ViewChild("search")
    private searchInput

    get searchEntry() : boolean {
        return this._searchEntry
    }

    set searchEntry(val: boolean) {
        this._searchEntry = val
        if (val) {
            setTimeout(() => {
                this.searchInput.nativeElement.focus()
            }, 10)
        }
    }

    get searchQuery() : string {
        return this._searchQuery
    }

    set searchQuery(val: string) {
        this._searchQuery = val
        localStorage.setItem(ToolbarFeedComponent.key, val)
    }

    private static key = "searchQuery"

    constructor(
        private articleService: ArticleService,
        private featuresServices: FeaturesService,
        private preferences : PreferencesService,
        private router: Router,
        private location: Location,
        private sharingService: SharingService,
    ) { 
        this.searchQuery = localStorage.getItem(ToolbarFeedComponent.key) || ""
    }

    ngOnInit(): void {
        let articleRouteObservable = articleRoute(this.router)

        this.subscriptions.push(articleRouteObservable.map(
            route => route != null
        ).subscribe(
            showsArticle => this.showsArticle = showsArticle
        ));

        this.articleID = articleRouteObservable.map(route => {
            if (route == null) {
                return -1;
            }

            return +route.params["articleID"];
        }).distinctUntilChanged().shareReplay(1)

        this.subscriptions.push(this.articleID.switchMap(id => {
            if (id == -1) {
                return Observable.of(false);
            }

            let initial = true
            return this.articleService.articleObservable().map(articles => {
                for (let article of articles) {
                    if (article.id == id) {
                        return article.read;
                    }
                }

                return false;
            }).delayWhen(read => {
                if (read && !initial) {
                    return Observable.timer(1000);
                }

                initial = false

                return Observable.timer(0);
            })
        }).subscribe(
            read => this.articleRead = read,
            error => console.log(error),
        ))

        this.subscriptions.push(listRoute(this.router).map(
            route => route != null && route.data["primary"] == "search"
        ).subscribe(
            inSearch => this.inSearch = inSearch,
            error => console.log(error),
        ))

        this.subscriptions.push(this.featuresServices.getFeatures().filter(
            features => features.search
        ).switchMap(features =>
            articleRouteObservable.map(
                route => route == null
            ).distinctUntilChanged().combineLatest(
                listRoute(this.router),
                (inList, route) : [boolean, boolean, boolean] => {
                    let showButton = false;
                    let showEntry = false;
                    let showAllRead = false;
                    if (inList) {
                        let route = getListRoute([this.router.routerState.snapshot.root])

                        switch (route.data["primary"]) {
                        case "favorite":
                            showAllRead = true;
                        case "popular":
                            break
                        case "search":
                            showEntry = true;
                        default:
                            showButton = true;
                            showAllRead = true;
                        }
                    }

                    return[showButton, showEntry, showAllRead] 
                }
            )
        ).subscribe(
            res => {
                this.searchButton = res[0]
                this.searchEntry = res[1]
                this.markAllRead = res[2]
            },
            error => console.log(error),
        ));

        this.subscriptions.push(this.sharingService.enabledServices().subscribe(
            services => {
                this.enabledShares = services.length > 0;
                this.shareServices = services;
            },
            error => console.log(error),
        ))
    }

    ngOnDestroy(): void {
        this.subscriptions.forEach(s => s.unsubscribe())
    }

    toggleOlderFirst() {
        this.preferences.olderFirst = !this.preferences.olderFirst;
        this.olderFirst = this.preferences.olderFirst;
    }

    toggleUnreadOnly() {
        this.preferences.unreadOnly = !this.preferences.unreadOnly;
    }

    markAsRead() {
        this.articleService.readAll();
    }

    up() {
        let path = this.location.path()
        let idx = path.indexOf("/article/")
        if (idx != -1) {
            this.router.navigateByUrl(path.substring(0, idx));
        } else if (this.inSearch) {
            this.router.navigateByUrl(path.substring(0, path.indexOf("/search/")));
        }
    }

    toggleRead() {
        this.articleID.take(1).switchMap(id => {
            if (id == -1) {
                Observable.empty();
            }

            return this.articleService.articleObservable().map(articles => {
                for (let article of articles) {
                    if (article.id == id) {
                        return article;
                    }
                }

                return null;
            }).take(1)
        }).flatMap(article =>
            this.articleService.read(article.id, !article.read)
        ).subscribe(
            success => {},
            error => console.log(error),
        )
    }

    keyEnter() {
        if (this.searchEntry && this.searchQuery) {
            this.performSearch(this.searchQuery);
        }
    }

    performSearch(query: string) {
        let route = getListRoute([this.router.routerState.snapshot.root])
        if (route.data["primary"] == "search") {
            let idx = this.location.path().indexOf("/search/") + 8
            this.router.navigateByUrl(this.location.path().substring(0, idx) + encodeURIComponent(query));
        } else {
            this.router.navigateByUrl(this.location.path() + "/search/" + encodeURIComponent(query));
        }
    }

    refresh() {
        this.articleService.refreshArticles()
    }

    shareArticleTo(share: ShareService) {
        this.articleID.take(1).filter(
            id => id != -1
        ).switchMap(id =>
             this.articleService.articleObservable().map(articles =>
                articles.filter(a => a.id == id)
            ).filter(
                articles => articles.length > 0
            ).map(articles => articles[0])
        ).take(1).subscribe(
            article => this.sharingService.submit(share.id, article)
        )
    }
}