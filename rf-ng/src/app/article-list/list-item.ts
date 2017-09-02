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
            data => {
                this.item.favorite = favor;
            },
            error => console.log(error)
        )
    }
}