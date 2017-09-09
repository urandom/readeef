import { Component, Input, OnInit } from '@angular/core';
import { ArticleService, Article } from "../services/article";

@Component({
    selector: "list-item",
    templateUrl: "./list-item.html",
    styleUrls: ["./list-item.css"],
})
export class ListItemComponent implements OnInit {
    @Input()
    item: Article

    constructor(private articleService: ArticleService) { }

    ngOnInit(): void {
        if (this.item.hits && this.item.hits.fragments) {
            if (this.item.hits.fragments.title.length > 0) {
                this.item.title = this.item.hits.fragments.title.join(" ")
            }

            if (this.item.hits.fragments.description.length > 0) {
                this.item.stripped = this.item.hits.fragments.description.join(" ")
            }
        }
    }

    favor(id: number, favor: boolean) {
        this.articleService.favor(id, favor).subscribe(
            success => { },
            error => console.log(error)
        )
    }
}