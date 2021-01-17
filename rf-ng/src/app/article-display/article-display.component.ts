import {
     Component, OnInit, OnDestroy,
     ViewChild, HostListener, ElementRef,
} from '@angular/core';
import { DomSanitizer } from "@angular/platform-browser";
import { ActivatedRoute, Router } from '@angular/router';
import { NgbCarouselConfig, NgbCarousel, NgbSlideEvent, NgbSlideEventSource } from '@ng-bootstrap/ng-bootstrap';
import { Article, ArticleFormat, ArticleService } from '../services/article'
import { FeaturesService } from '../services/features'
import { Observable, Subscription, Subject, BehaviorSubject, of, interval, from } from 'rxjs';
import * as moment from 'moment';
import { switchMap, startWith, map, filter, distinctUntilChanged, mergeMap, catchError, ignoreElements, take } from 'rxjs/operators';
import { InteractionService } from '../services/interaction';

enum State {
    DESCRIPTION, FORMAT, SUMMARY,
}

interface NavigationPayload {
    slides: Article[],
    active: {
        id: number,
        read: boolean,
        state: State,
    },
    index: number,
    total: number,
}

@Component({
    selector: "article-display",
    templateUrl: "./article-display.html",
    styleUrls: ["./article-display.css"],
    providers: [ NgbCarouselConfig ],
})
export class ArticleDisplayComponent implements OnInit, OnDestroy {
    canExtract: boolean;

    slides: Article[] = [];
    index: string;
    activeId: string;

    @ViewChild('carousel', {static: false})
    private carousel: NgbCarousel;

    @ViewChild('carousel', {read: ElementRef})
    private carouselEl: ElementRef;

    private active: Article;
    private offset = new Subject<number>();
    private stateChange = new BehaviorSubject<[number, State]>([-1, State.DESCRIPTION]);
    private states = new Map<number, State>();
    private subscriptions = new Array<Subscription>();

    constructor(
        config: NgbCarouselConfig,
        private route: ActivatedRoute,
        private router: Router,
        private articleService: ArticleService,
        private featuresService: FeaturesService,
        private sanitizer: DomSanitizer,
        private interactionService: InteractionService,
    ) {
        config.interval = 0;
        config.wrap = false;
        config.keyboard = false;
        config.showNavigationArrows = false;
    }

    ngOnInit(): void {
        this.subscriptions.push(this.articleService.articleObservable().pipe(
            filter(articles => articles !== true),
            map(articles => (articles as Article[])),
            switchMap(articles =>
                this.stateChange.pipe(
                    switchMap(stateChange =>
                        this.offset.pipe(
                            startWith(0),
                            map((offset): [number, [number, State]] => {
                                return [offset, stateChange]
                            })
                        ),
                    ),
                    map((offsetState): NavigationPayload => {
                        let [offset, stateChange] = offsetState
                        let id = this.route.snapshot.params["articleID"];
                        let slides: Article[] = [];
                        let index = articles.findIndex(article => article.id == id)

                        if (index == -1) {
                            return null
                        }

                        if (offset != 0) {
                            if (index + offset != -1 && index + offset < articles.length) {
                                index += offset;
                            }

                            this.router.navigate(['../', articles[index].id], { relativeTo: this.route })
                        }

                        if (index > 0) {
                            slides.push(articles[index - 1]);
                        }
                        slides.push(articles[index]);
                        if (index + 1 < articles.length) {
                            slides.push(articles[index + 1]);
                        }

                        slides = slides.map(slide => {
                            if (slide.id == stateChange[0]) {
                                switch (stateChange[1]) {
                                    case State.DESCRIPTION:
                                        slide["formatted"] = slide.description
                                        break
                                    case State.FORMAT:
                                        slide["formatted"] = slide.format.content
                                        break
                                    case State.SUMMARY:
                                        slide["formatted"] = this.keypointsToHTML(slide.format)
                                        break
                                }
                            } else {
                                slide["formatted"] = slide.description
                            }

                            slide["formatted"] =
                                this.sanitizer.bypassSecurityTrustHtml(
                                    this.formatSource(slide["formatted"])
                                )
                            return slide
                        })

                        return {
                            slides: slides,
                            active: {
                                id: articles[index].id,
                                read: articles[index].read,
                                state: stateChange[1],
                            },
                            index: index,
                            total: articles.length,
                        }
                    }),
                )
            ),
            filter(data => data != null),
            distinctUntilChanged((a, b) =>
                a.active.id == b.active.id && a.slides.length == b.slides.length &&
                a.active.state == b.active.state && a.total == b.total &&
                (a.slides[0] || {})['id'] == (b.slides[0] || {})['id'] &&
                (a.slides[2] || {})['id'] == (b.slides[2] || {})['id']
            ),
            mergeMap(data => {
                if (data.active.read) {
                    return of(data);
                }

                return this.articleService.read(data.active.id, true).pipe(
                    map(_ => data),
                    catchError(err => of(err)),
                    ignoreElements(),
                    startWith(data),
                );
            }),
            switchMap(data => interval(60000).pipe(
                startWith(0),
                map(_ => {
                    data.slides = data.slides.map(article => {
                        article.time = moment(article.date).fromNow();
                        return article;
                    })

                    return data;
                }),
            ))
        ).subscribe(
            data => {
                this.slides = data.slides;
                this.active = data.slides.find(article => article.id == data.active.id);

                if (data.slides.length == 2 && data.slides[1].id == data.active.id) {
                    this.articleService.requestNextPage()
                }

                this.index = `${data.index + 1}/${data.total}`
                this.activeId = data.active.id.toString();

                if (this.carousel) {
                    this.carousel.select(data.active.id.toString(), NgbSlideEventSource.TIMER);
                }
            },
            error => console.log(error)
        ));

        this.subscriptions.push(this.featuresService.getFeatures().pipe(
            map(features => features.extractor),
        ).subscribe(
            canExtract => this.canExtract = canExtract,
            error => console.log(error),
        ));

        this.subscriptions.push(this.stateChange.subscribe(
            stateChange => this.states.set(stateChange[0], stateChange[1])
        ));

        this.subscriptions.push(this.interactionService.toolbarTitleClickEvent.subscribe(
            () => {
                this.carouselEl.nativeElement.scrollTop = 0;
            },
        ));
    }

    ngOnDestroy(): void {
        for (let subscription of this.subscriptions) {
            subscription.unsubscribe()
        }
    }

    slideEvent(event: NgbSlideEvent) {
        if (event.source == NgbSlideEventSource.TIMER) {
            return;
        }

        if (event.direction == "left") {
            this.offset.next(1);
        } else {
            this.offset.next(-1);
        }

        this.carouselEl.nativeElement.scrollTop = 0;
    }

    swipeEvent(next: boolean) {
        if (next) {
            this.carousel.next()
        } else {
            this.carousel.prev()
        }
    }

    favor(id: number, favor: boolean) {
        this.articleService.favor(id, favor).subscribe(
            _ => { },
            error => console.log(error)
        )
    }

    @HostListener('window:keydown.Escape')
    @HostListener('window:keydown.shift.arrowUp')
    @HostListener('window:keydown.h')
    goUp() {
        this.router.navigate(['../../'], { relativeTo: this.route, state: {"articleID": this.active.id} })
        document.getSelection().removeAllRanges();
    }

    @HostListener('window:keydown.arrowRight')
    @HostListener('window:keydown.j')
    goNext() {
        this.carousel.next()
    }

    @HostListener('window:keydown.arrowLeft')
    @HostListener('window:keydown.k')
    goPrevious() {
        this.carousel.prev()
    }

    @HostListener('window:keydown.shift.arrowLeft')
    @HostListener('window:keydown.shift.j')
    previousUnread() {
        let id = this.route.snapshot.params["articleID"];
        this.articleService.articleObservable().pipe(
            filter(articles => articles !== true),
            map(articles => (articles as Article[])),
            map(articles => {
                let idx = articles.findIndex(article => article.id == id);

                while (idx > 0) {
                    idx--;
                    if (!articles[idx].read) {
                        break;
                    }
                }

                if (articles[idx].read) {
                    return id;
                }

                return articles[idx].id;
            }),
            take(1),
            filter(a => a != id),
            mergeMap(id =>
                from(this.router.navigate(
                    ['../', id], { relativeTo: this.route }
                )).pipe(map(_ => id))
        )).subscribe(
            id => {
                this.stateChange.next([id, State.DESCRIPTION]);
            }
        )
    }

    @HostListener('window:keydown.shift.arrowRight')
    @HostListener('window:keydown.shift.k')
    nextUnread() {
        let id = this.route.snapshot.params["articleID"];
        this.articleService.articleObservable().pipe(
            filter(articles => articles !== true),
            map(articles => (articles as Article[])),
            map(articles => {
                let idx = articles.findIndex(article => article.id == id);

                while (idx < articles.length - 1) {
                    idx++;
                    if (!articles[idx].read) {
                        break;
                    }
                }

                if (articles[idx].read) {
                    return id;
                }

                return articles[idx].id;
            }),
            take(1),
            filter(a => a != id),
            mergeMap(id =>
                from(this.router.navigate(
                    ['../', id], { relativeTo: this.route }
                )).pipe(map(_ => id))
            ),
        ).subscribe(
            id => {
                this.stateChange.next([id, State.DESCRIPTION]);
            }
        )
    }

    @HostListener('window:keydown.v')
    viewActive() : boolean {
        if (this.active != null) {
            if (document.body.dispatchEvent(new CustomEvent('open-link', {
                cancelable: true,
                detail: this.active.link,
            }))) {
                window.open(this.active.link, "_blank");
            }
        }

        return false;
    }

    @HostListener('window:keydown.c')
    formatActive() {
        if (this.active != null) {
            this.formatArticle(this.active)
        }
    }

    formatArticle(article: Article) {
        let state = this.getState(article.id)
        if (state == State.FORMAT) {
            state = State.DESCRIPTION
        } else {
            state = State.FORMAT
        }

        this.setFormat(article, state)
    }

    @HostListener('window:keydown.s')
    summarizeActive() {
        if (this.active != null) {
            this.summarizeArticle(this.active)
        }
    }

    summarizeArticle(article: Article) {
        let state = this.getState(article.id)
        if (state == State.SUMMARY) {
            state = State.DESCRIPTION
        } else {
            state = State.SUMMARY
        }

        this.setFormat(article, state)
    }

    @HostListener('window:keydown.f')
    favorActive() {
        if (this.active != null) {
            this.favorArticle(this.active)
        }
    }

    favorArticle(article: Article) {
        this.articleService.favor(
            article.id, !article.favorite
        ).subscribe(
            _ => { },
            error => console.log(error)
        )
    }

    private keypointsToHTML(format: ArticleFormat) : string {
        return `<img src="${format.topImage}" class="top-image center"><br><ul><li>`
            + format.keyPoints.join("<li>")
            + "</ul>"
    }

    private getState(id: number) : State {
        if (this.states.has(id)) {
            return this.states.get(id)
        }

        return State.DESCRIPTION
    }

    private setFormat(article: Article, state: State) {
        let active = this.active
        if (state == State.DESCRIPTION) {
            this.stateChange.next([active.id, state])
            return
        }

        let o : Observable<ArticleFormat>
        if (article.format) {
            o = of(article.format)
        } else {
            o = this.articleService.formatArticle(active.id)
        }

        o.subscribe(
            format => {
                active.format = format
                this.stateChange.next([active.id, state])
            },
            error => console.log(error)
        )
    }

    private formatSource(source: string) : string {
        return source.replace("<img ", `<img class="center" `);
    }
}
