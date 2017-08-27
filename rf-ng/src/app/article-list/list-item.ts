import { Component, Input } from '@angular/core';

export interface ListItem {
    title: string,
    description: string,
    thumbnail: string,
    feed?: string,
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
}