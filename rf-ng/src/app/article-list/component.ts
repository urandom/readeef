import { Component, OnInit, OnDestroy, HostListener } from "@angular/core";
import { Router, ActivatedRoute, ParamMap, Data, Params } from '@angular/router';
import { Article, Source, UserSource, FavoriteSource, PopularSource, FeedSource, TagSource, ArticleService, QueryOptions } from "../services/article"
import { ChangeEvent } from 'angular2-virtual-scroll';
import { Observable, Subscription } from "rxjs";
import { BehaviorSubject } from "rxjs/BehaviorSubject";
import * as moment from 'moment';
import 'rxjs/add/observable/interval'
import 'rxjs/add/operator/scan'
import 'rxjs/add/operator/startWith'
import 'rxjs/add/operator/switchMap'

class ArticleCounter {
    constructor(
        public iteration: number,
        public articles: Array<Article>,
    ) { }
}

@Component({
    selector: "article-list",
    templateUrl: "./article-list.html",
    styleUrls: ["./article-list.css"],
})
export class ArticleListComponent implements OnInit, OnDestroy {
    items: Article[] = []
    scrollItems: Article[]
    loading: boolean

    private finished = false
    private limit: number = 200;
    private subscription: Subscription;

    constructor(
        private articleService: ArticleService,
        private router: Router,
        private route: ActivatedRoute,
    ) {
    }

    ngOnInit(): void {
        this.loading = true;

        this.subscription = this.articleService.articleObservable(
        ).scan((acc, articles) => {
            if (acc.iteration > 0 && acc.articles.length == articles.length) {
                this.finished = true
            }

            acc.articles = [].concat(articles)
            acc.iteration++
            return acc
        }, new ArticleCounter(0, [])).map(
            acc => acc.articles
        ).switchMap(articles =>
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
        if (event.end == this.items.length && !this.loading && !this.finished) {
            this.loading = true;
            this.articleService.requestNextPage();
        }
    }

    @HostListener('window:keydown.arrowLeft')
    @HostListener('window:keydown.shift.j')
    firstUnread() {
        if (document.activeElement.matches("input")) {
            return
        }
        let article = this.items.find(article => !article.read)
        if (article) {
            this.router.navigate(['article', article.id], {relativeTo: this.route})
        }
    }

    @HostListener('window:keydown.arrowRight')
    @HostListener('window:keydown.shift.k')
    lastUnread() {
        if (document.activeElement.matches("input")) {
            return
        }
        for (let i = this.items.length - 1; i > -1; i--) {
            let article = this.items[i]
            if (!article.read) {
                this.router.navigate(['article', article.id], {relativeTo: this.route})
                return
            }
        }
    }

    @HostListener('window:keydown.r')
    refresh() {
        if (document.activeElement.matches("input")) {
            return
        }
        this.articleService.refreshArticles()
    }
}