import { Component, Input } from '@angular/core';
import { ArticleService } from "../services/article";

export interface ListItem {
    id: number,
    title: string,
    description: string,
    thumbnail: string,
    feed: string,
    date: Date,
    time: string,
    read: boolean,
    favorite: boolean,
}

@Component({
    selector: "list-item",
    templateUrl: "./list-item.html",
    styleUrls: ["./list-item.css"],
})
export class ListItemComponent {
    @Input()
    item: ListItem

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