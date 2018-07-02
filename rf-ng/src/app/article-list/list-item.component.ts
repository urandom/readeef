import { Component, Input } from '@angular/core';
import { Router, ActivatedRoute } from "@angular/router";
import { ArticleService, Article } from "../services/article";

@Component({
    selector: "list-item",
    templateUrl: "./list-item.html",
    styleUrls: ["./list-item.css"],
})
export class ListItemComponent {
    @Input()
    item: Article

    constructor(
        private articleService: ArticleService,
        private router: Router,
        private route: ActivatedRoute,
    ) { }

    openArticle(article: Article) {
        this.router.navigate(['article', article.id], {relativeTo: this.route});
    }

    favor(id: number, favor: boolean) {
        this.articleService.favor(id, favor).subscribe(
            success => { },
            error => console.log(error)
        )
    }
}