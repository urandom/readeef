import { Component, OnInit, OnDestroy } from "@angular/core";
import { ActivatedRoute, ParamMap, Data, Params } from '@angular/router';
import { Article, Source, UserSource, FavoriteSource, PopularSource, FeedSource, TagSource, ArticleService, QueryOptions } from "../services/article"
import { ChangeEvent } from 'angular2-virtual-scroll';
import { Observable, Subscription } from "rxjs";
import { BehaviorSubject } from "rxjs/BehaviorSubject";
import * as moment from 'moment';
import 'rxjs/add/observable/interval'
import 'rxjs/add/operator/startWith'
import 'rxjs/add/operator/switchMap'

@Component({
    selector: "article-list",
    templateUrl: "./article-list.html",
    styleUrls: ["./article-list.css"],
})
export class ArticleListComponent implements OnInit, OnDestroy {
    items: Article[] = []
    loading: boolean

    private limit: number = 200;
    private subscription: Subscription;

    constructor(
        private articleService: ArticleService,
        private route: ActivatedRoute,
    ) {
    }

    ngOnInit(): void {
        this.loading = true;

        this.subscription = this.articleService.articleObservable(
        ).startWith([]).switchMap(articles =>
            Observable.interval(60000).startWith(0).map(v =>
                articles.map(article => {
                    article.time = moment(article.date).fromNow();
                    return article;
                })
            )
        ).subscribe(
            articles => {
                this.loading = false;
                this.items = articles;
            },
            error => {
                this.loading = false;
                console.log(error);
            }
        )
    }

    ngOnDestroy(): void {
        this.subscription.unsubscribe();
    }

    fetchMore(event: ChangeEvent) {
        if (event.end == this.items.length && !this.loading) {
            this.loading = true;
            this.articleService.requestNextPage();
        }
    }
}