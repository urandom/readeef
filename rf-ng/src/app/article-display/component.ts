import { Component, OnInit, OnDestroy, Input, ViewChild, ElementRef, OnChanges } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { Location } from '@angular/common';
import { Article, ArticleFormat, ArticleService } from '../services/article'
import { FeaturesService } from '../services/features'
import { Observable, Subscription, Subject, BehaviorSubject } from 'rxjs';
import 'rxjs/add/observable/of'
import 'rxjs/add/operator/combineLatest'
import 'rxjs/add/operator/distinctUntilChanged'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/switchMap'
import { NgbCarouselConfig, NgbCarousel } from '@ng-bootstrap/ng-bootstrap';

enum State {
    DESCRIPTION, FORMAT, SUMMARY,
}

@Component({
    selector: "article-display",
    templateUrl: "./article-display.html",
    styleUrls: ["./article-display.css"],
    providers: [ NgbCarouselConfig ],
    host: {
        '(keydown.arrowUp)': 'goUp()',
        '(keydown.v)': 'viewActive()',
        '(keydown.c)': 'formatActive()',
        '(keydown.s)': 'summarizeActive()',
        '(keydown.f)': 'favorActive()',
    }
})
export class ArticleDisplayComponent implements OnInit, OnDestroy {
    canExtract: boolean;
    @Input()
    items: Article[] = []

    slides: Article[] = []

    @ViewChild("carousel")
    private carousel : NgbCarousel;
    @ViewChild("carousel", {read: ElementRef})
    private carouselElement: ElementRef;

    private active: Article
    private offset = new Subject<number>()
    private stateChange = new BehaviorSubject<[number, State]>([-1, State.DESCRIPTION])
    private states = new Map<number, State>();
    private subscriptions = new Array<Subscription>()

    constructor(
        config: NgbCarouselConfig,
        private route: ActivatedRoute,
        private router: Router,
        private location: Location,
        private articleService: ArticleService,
        private featuresService: FeaturesService,
    ) {
        config.interval = 0
        config.wrap = false
        config.keyboard = true
    }

    ngOnInit(): void {
        this.subscriptions.push(this.articleService.articleObservable(
        ).switchMap(articles =>
            this.stateChange.flatMap(stateChange => 
                this.offset.startWith(0).map(
                    (offset) : [number, [number | State]] => {
                        return [offset, stateChange]
                    }
                )
            ).map(
                (offsetState): [Article[], number, boolean] => {
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

                        let path = this.location.path();
                        path = path.substring(0, path.lastIndexOf("/") + 1) + articles[index].id;
                        this.router.navigateByUrl(path)
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
                        return slide
                    })

                    return [slides, articles[index].id, articles[index].read];
                }
            )
        ).filter(
            data => data != null
        ).distinctUntilChanged((a, b) =>
            a[1] == b[1] && a[0].length == b[0].length
        ).flatMap(data => {
            if (data[2]) {
                return Observable.of(data);
            }
            return this.articleService.read(data[1], true).map(s => data);
        }).subscribe(
            data => {
                let [slides, id] = data
                this.carousel.activeId = id.toString();
                this.slides = slides;
                this.active = slides.find(article => article.id == id);

                if (slides.length == 2 && slides[1].id == id) {
                    this.articleService.requestNextPage()
                }

                setTimeout(() => {
                    this.stylizeContent(this.carouselElement.nativeElement)
                }, 5)
            },
            error => console.log(error)
        ));

        this.subscriptions.push(this.featuresService.getFeatures().map(
            features => features.extractor
        ).subscribe(
            canExtract => this.canExtract = canExtract,
            error => console.log(error),
        ))

        this.subscriptions.push(this.stateChange.subscribe(
            stateChange => this.states.set(stateChange[0], stateChange[1])
        ))

        this.carouselElement.nativeElement.focus()
    }

    ngOnDestroy(): void {
        for (let subscription of this.subscriptions) {
            subscription.unsubscribe()
        }
    }

    slideEvent(next: boolean) {
        if (next) {
            this.offset.next(1);
        } else {
            this.offset.next(-1);
        }
    }

    favor(id: number, favor: boolean) {
        this.articleService.favor(id, favor).subscribe(
            success => { },
            error => console.log(error)
        )
    }

    goUp() {
        let path = this.location.path();
        this.router.navigateByUrl(path.substring( 0, path.indexOf("/article/")))
    }

    viewActive() {
        if (this.active != null) {
            window.open(this.active.link, "_blank");
        }
    }

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

    favorActive() {
        if (this.active != null) {
            this.favorArticle(this.active)
        }
    }

    favorArticle(article: Article) {
        this.articleService.favor(
            article.id, !article.favorite
        ).subscribe(
            success => { },
            error => console.log(error)
        )
    }

    private keypointsToHTML(format: ArticleFormat) : string {
        return `<img src="${format.topImage}" class="top-image"><br><ul><li>`
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
            o = Observable.of(article.format)
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

    private stylizeContent(container) {
        let styler = (img) => {
            if (img.naturalWidth < img.parentNode.offsetWidth * 0.666) {
                if (img.parentNode.textContent.length < 100) {
                    img.classList.add("center")
                } else {
                    img.classList.add("float")

                    if (!img.parentNode.__clearAdded) {
                        let clear = document.createElement('br')
                        clear.style.clear = 'both'
                        img.parentNode.appendChild(clear)
                        img.parentNode.__clearAdded = true
                    }
                }
            }
        }

        for (let descr of container.querySelectorAll(".description")) {
            for (let img of descr.querySelectorAll("img")) {
                if (img.complete) {
                    styler(img)
                } else {
                    img.addEventListener('load', () => styler(img))
                }
            }
        }
    }
}