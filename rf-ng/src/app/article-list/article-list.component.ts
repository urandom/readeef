import { Component, OnInit, OnDestroy, HostListener, ViewChild } from "@angular/core";
import { Router, ActivatedRoute } from '@angular/router';;
import { Article, ArticleService } from "../services/article";
import { IPageInfo, VirtualScrollerComponent } from 'ngx-virtual-scroller';
import { Subscription, interval } from "rxjs";
import * as moment from 'moment';
import { scan, map, switchMap, startWith } from "rxjs/operators";
import { InteractionService } from '../services/interaction';
import { PreferencesService } from '../services/preferences';

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
    items: Article[] = [];
    loading: boolean;

    private finished = false
    private subscriptions = new Array<Subscription>();
    private articleID: number;

    @ViewChild('scroll', {static: true})
    private scroller: VirtualScrollerComponent;

    constructor(
        private articleService: ArticleService,
        private router: Router,
        private route: ActivatedRoute,
        private interactionService: InteractionService,
        private preferencesService: PreferencesService,
    ) {
        this.articleID = (this.router.getCurrentNavigation().extras.state ?? {})["articleID"];
    }

    ngOnInit(): void {
        this.loading = true;

        this.subscriptions.push(this.articleService.articleObservable().pipe(
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
                if (this.articleID || this.loading) {
                    setTimeout(() => {
                        if (this.loading) {
                            this.scroller.scrollToIndex(0);
                            return;
                        }
                        let idx = this.items.findIndex(i => i.id == this.articleID);
                        if (idx != -1) {
                            this.scroller.scrollToIndex(idx);
                        }
                        this.articleID = 0;
                    }, 0);
                }
            },
            error => {
                console.log(error);
                this.loading = false;
            }
        ));

        this.subscriptions.push(this.interactionService.toolbarTitleClickEvent.subscribe(
            () => {
                if (this.scroller.viewPortInfo.startIndex == 0) {
                    // If already at the start of the list, go to the oldest unread item

                    let idx = -1;
                    if (this.preferencesService.olderFirst) {
                        for (let i = 0; i < this.items.length; i++) {
                            if (!this.items[i].read) {
                                idx = i;
                                break;
                            }       
                        }
                    } else {
                        for (let i = this.items.length - 1; i > -1; i--) {
                            if (!this.items[i].read) {
                                idx = i;
                                break;
                            }
                        }
                    }

                    if (idx != -1) {
                        this.scroller.scrollToIndex(idx);
                    }
                    return;
                }
                this.scroller.scrollToIndex(0);
            },
        ));
    }

    ngOnDestroy(): void {
        this.subscriptions.forEach(s => s.unsubscribe());
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
            return;
        }
        this.articleService.refreshArticles();
    }
}
