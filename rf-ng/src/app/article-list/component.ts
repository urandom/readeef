import { Component, OnInit, OnChanges, Input, SimpleChanges } from "@angular/core";
import { ActivatedRoute, ParamMap } from '@angular/router';
import { Article, ArticleService } from "../services/article"
import { ListItem } from './list-item';
import { ChangeEvent } from 'angular2-virtual-scroll';
import { Observable } from "rxjs";

@Component({
    selector: "article-list",
    templateUrl: "./article-list.html",
    styleUrls: ["./article-list.css"],
})
export class ArticleListComponent implements OnChanges {
    @Input()
    items: ListItem[] = []
    limit: number = 200
    loading: boolean

    constructor(
        private articleService: ArticleService,
        private route: ActivatedRoute,
    ) {}

    ngOnChanges(changes: SimpleChanges): void {
    }

    fetchMore(event: ChangeEvent) {
        if (event.end == this.items.length) {
            this.loading = true;
            this.getArticles(this.limit, this.items.length).subscribe(
                articles => {
                    this.loading = false;
                    console.log(articles);
                },
                error => {
                    this.loading = false;
                    console.log(error);
                }
            );
        }
    }

    private getArticles(limit: number, offset: number) : Observable<Article[]> {
        return this.articleService.getArticles();
    }
}