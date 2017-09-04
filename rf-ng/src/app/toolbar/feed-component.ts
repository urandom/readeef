import { Component, OnInit, OnDestroy } from '@angular/core';
import { Location } from '@angular/common';
import { Article, ArticleService } from "../services/article";
import { PreferencesService } from "../services/preferences";
import { Router, NavigationStart } from '@angular/router';
import { Observable, Subscription } from "rxjs";
import { articleDisplayRoute, articlePattern } from "../main/routing-util"
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
        private location: Location,
    ) { }

    ngOnInit(): void {
        this.subscriptions.push(articleDisplayRoute(
            this.router, this.location
        ).subscribe(
            showsArticle => this.showsArticle = showsArticle
        ));

        this.articleID = this.router.events.map(event => {
            if (event instanceof NavigationStart) {
                return event.url;
            }

            return "";
        }).filter(path => path != "").startWith(
            this.location.path()
        ).map(path => {
            let idx = path.indexOf(articlePattern)
            if (idx == -1) {
                return -1;
            }

            path = path.substring(idx + articlePattern.length);
            idx = path.indexOf("/");
            if (idx != -1) {
                path = path.substring(0, idx);
            }

            return +path;
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
        let idx = this.location.path().indexOf("/article/");
        if (idx != -1) {
            this.router.navigateByUrl(this.location.path().substring(0, idx));
        }
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