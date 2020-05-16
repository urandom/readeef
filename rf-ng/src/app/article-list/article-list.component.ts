import { Component, OnInit, OnDestroy, HostListener } from "@angular/core";
import { Router, ActivatedRoute } from '@angular/router';;
import { Article, ArticleService } from "../services/article";
import { IPageInfo } from 'ngx-virtual-scroller';
import { Subscription, interval } from "rxjs";
import * as moment from 'moment';
import { scan, filter, map, switchMap, startWith } from "rxjs/operators";

class ArticleCounter {
    constructor(
        public loading: boolean,
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
    private subscription: Subscription;

    constructor(
        private articleService: ArticleService,
        private router: Router,
        private route: ActivatedRoute,
    ) {
    }

    ngOnInit(): void {
        this.loading = true;

        this.subscription = this.articleService.articleObservable().pipe(
            scan<Article[]|true, ArticleCounter>((acc, articles, _) => {
                if (articles === true) { 
                    acc.loading = true;
                    return acc;
                }

                if (acc.iteration > 0 && acc.articles.length == articles.length) {
                    this.finished = true;
                }

                acc.articles = [].concat(articles);
                acc.iteration++;
                acc.loading = false;
                return acc;
            }, new ArticleCounter(false, 0, [])),
            switchMap(acc => interval(60000).pipe(
                startWith(0),
                map(_ => {
                    acc.articles.map(article => {
                        article.time = moment(article.date).fromNow();
                        return article;
                    });
                    return acc;
                }),
            )
        )).subscribe(
            acc => {
                this.items = acc.articles;
                this.loading = acc.loading;
            },
            error => {
                console.log(error);
                this.loading = false;
            }
        )
    }

    ngOnDestroy(): void {
        this.subscription.unsubscribe();
    }

    fetchMore(event: IPageInfo) {
        if (this.items.length > 0 && event.endIndex === this.items.length - 1 && !this.loading && !this.finished) {
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
