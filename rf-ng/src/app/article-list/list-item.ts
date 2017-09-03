import { Component, Input } from '@angular/core';
import { ArticleService, Article } from "../services/article";

@Component({
    selector: "list-item",
    templateUrl: "./list-item.html",
    styleUrls: ["./list-item.css"],
})
export class ListItemComponent {
    @Input()
    item: Article

    constructor(private articleService: ArticleService) { }

    favor(id: number, favor: boolean) {
        this.articleService.favor(id, favor).subscribe(
            success => { },
            error => console.log(error)
        )
    }
}