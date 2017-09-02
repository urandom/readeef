import { Component, OnInit, Input } from "@angular/core";
import { ActivatedRoute } from "@angular/router";
import { Article, ArticleService } from "../services/article"
import { Observable } from "rxjs";
import { NgbCarouselConfig } from '@ng-bootstrap/ng-bootstrap';

@Component({
    selector: "article-display",
    templateUrl: "./article-display.html",
    styleUrls: ["./article-display.css"],
    providers: [ NgbCarouselConfig ],
})
export class ArticleDisplayComponent implements OnInit {
    @Input()
    items: Article[] = []

    slides: Article[] = []

    constructor(
        config: NgbCarouselConfig,
        private route: ActivatedRoute,
        private articleService: ArticleService,
    ) {
        config.interval = 0
        config.wrap = true
        config.keyboard = true
    }

    ngOnInit(): void {
    }

}