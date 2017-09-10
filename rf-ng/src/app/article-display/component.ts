import { Component, OnInit, OnDestroy, Input, ViewChild, ElementRef } from "@angular/core";
import { ActivatedRoute, Router } from "@angular/router";
import { Location } from '@angular/common';
import { Article, ArticleService } from "../services/article"
import { Observable, Subscription } from "rxjs";
import { Subject } from "rxjs/Subject";
import 'rxjs/add/observable/of'
import 'rxjs/add/operator/distinctUntilKeyChanged'
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
    }
})
export class ArticleDisplayComponent implements OnInit, OnDestroy {
    @Input()
    items: Article[] = []

    slides: Article[] = []

    @ViewChild("carousel")
    private carousel : NgbCarousel;
    @ViewChild("carousel", {read: ElementRef})
    private carouselElement: ElementRef;

    private offset = new Subject<number>();
    private subscription: Subscription;

    constructor(
        config: NgbCarouselConfig,
        private route: ActivatedRoute,
        private router: Router,
        private location: Location,
        private articleService: ArticleService,
    ) {
        config.interval = 0
        config.wrap = false
        config.keyboard = true
    }

    ngOnInit(): void {
        this.subscription = this.articleService.articleObservable(
        ).switchMap(articles =>
            this.offset.startWith(0).map((offset) : [Article[], number, boolean] => {
                let id = this.route.snapshot.params["articleID"];
                let index = -1
                let slides : Article[] = [];

                articles.some((article, idx) => {
                    if (article.id == id) {
                        index = idx;
                        return true;
                    }
                    return false;
                });

                if (index + offset != -1 && index + offset < articles.length) {
                    index += offset;
                }

                if (offset != 0) {
                    let path = this.location.path();
                    path = path.substring(0, path.lastIndexOf("/") + 1) + articles[index].id;
                    this.router.navigateByUrl(path)
                }

                if (index != -1) {
                    if (index > 0) {
                        slides.push(articles[index-1]);
                    }
                    slides.push(articles[index]);
                    if (index + 1 < articles.length) {
                        slides.push(articles[index + 1]);
                    }
                }

                return [slides, articles[index].id, articles[index].read];
            })
        ).distinctUntilKeyChanged("1").flatMap(data => {
            if (data[2]) {
                return Observable.of(data);
            }
            return this.articleService.read(data[1], true).map(s => data);
        }).subscribe(
            data => {
                this.carousel.activeId = data[1].toString();
                this.slides = data[0];
            },
            error => console.log(error)
        );

        this.carouselElement.nativeElement.focus()
    }

    ngOnDestroy(): void {
        this.subscription.unsubscribe();
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
}