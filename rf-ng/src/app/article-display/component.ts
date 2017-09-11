import { Component, OnInit, OnDestroy, Input, ViewChild, ElementRef, OnChanges } from "@angular/core";
import { ActivatedRoute, Router } from "@angular/router";
import { Location } from '@angular/common';
import { Article, ArticleService } from "../services/article"
import { FeaturesService } from "../services/features"
import { Observable, Subscription } from "rxjs";
import { Subject } from "rxjs/Subject";
import 'rxjs/add/observable/of'
import 'rxjs/add/operator/distinctUntilChanged'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/switchMap'
import { NgbCarouselConfig, NgbCarousel } from '@ng-bootstrap/ng-bootstrap';

@Component({
    selector: "article-display",
    templateUrl: "./article-display.html",
    styleUrls: ["./article-display.css"],
    providers: [ NgbCarouselConfig ],
    host: {
        '(keydown.arrowUp)': 'keyUp()',
        '(keydown.v)': 'keyView()',
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
    private offset = new Subject<number>();
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
            this.offset.startWith(0).map((offset) : [Article[], number, boolean] => {
                let id = this.route.snapshot.params["articleID"];
                let slides : Article[] = [];
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
                    slides.push(articles[index-1]);
                }
                slides.push(articles[index]);
                if (index + 1 < articles.length) {
                    slides.push(articles[index + 1]);
                }

                return [slides, articles[index].id, articles[index].read];
            })
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

    keyUp() {
        let path = this.location.path();
        this.router.navigateByUrl(path.substring( 0, path.indexOf("/article/")))
    }

    keyView() {
        if (this.active != null) {
            window.open(this.active.link, "_blank");
        }
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