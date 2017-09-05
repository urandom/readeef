import { Component, OnInit, OnDestroy } from '@angular/core';
import { Article, ArticleService } from "../services/article";
import { PreferencesService } from "../services/preferences";
import { Router, NavigationStart } from '@angular/router';
import { Observable, Subscription } from "rxjs";
import { articleRoute, getListRoute } from "../main/routing-util"
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
  styleUrls: ['./toolbar.css']
})
export class ToolbarFeedComponent implements OnInit, OnDestroy {
    olderFirst = false
    showsArticle = false
    articleRead = false

    private articleID : Observable<number>
    private subscriptions = new Array<Subscription>()

    constructor(
        private articleService: ArticleService,
        private preferences : PreferencesService,
        private router: Router,
    ) { }

    ngOnInit(): void {
        this.subscriptions.push(articleRoute(this.router).map(
            route => route != null
        ).subscribe(
            showsArticle => this.showsArticle = showsArticle
        ));

        this.articleID = articleRoute(this.router).map(route => {
            if (route == null) {
                return -1;
            }

            return +route.params["articleID"];
        }).distinctUntilChanged().shareReplay(1)

        this.subscriptions.push(this.articleID.switchMap(id => {
            if (id == -1) {
                return Observable.of(false);
            }

            return this.articleService.articleObservable().map(articles => {
                for (let article of articles) {
                    if (article.id == id) {
                        return article.read;
                    }
                }

                return false;
            }).delayWhen(read => {
                if (read) {
                    return Observable.timer(1000);
                }

                return Observable.timer(0);
            })
        }).subscribe(
            read => this.articleRead = read,
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

    up() {
        let route = getListRoute([this.router.routerState.snapshot.root]);
        this.router.navigate([route]);
    }

    toggleRead() {
        this.articleID.switchMap(id => {
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
}