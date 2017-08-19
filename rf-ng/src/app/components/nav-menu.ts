import { Component, OnInit } from '@angular/core';
import { Tag, TagFeedIDs, TagService } from '../services/tag';
import { Feed, FeedService } from '../services/feed';
import { Features, FeaturesService } from "../services/features"
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/combineLatest'

@Component({
  selector: 'nav-menu',
  templateUrl: './nav-menu.html',
  styleUrls: ['./nav-menu.css']
})
export class NavMenuComponent implements OnInit {
    popularity: boolean
    collapses: Map<any, boolean> = new Map()
    popularityItems : Array<Item> = new Array();
    allItems : Array<Item> = new Array();
    tags : Array<Category> = new Array();

    constructor(
        private tagService: TagService,
        private feedService: FeedService,
        private featuresService: FeaturesService,
    ) {
        this.collapses.set("__popularity", true);
        this.collapses.set("__all", true);
    }

    ngOnInit(): void {
        this.featuresService.getFeatures().combineLatest(
            this.feedService.getFeeds(),
            this.tagService.getTagsFeedIDs(),
            (features, feeds, tags) => {
                let res : [Features, Feed[], TagFeedIDs[]];
                res = [features, feeds, tags];
                return res;
            }
        ).subscribe(
            data => {
                this.popularity = data[0].popularity;

                let feeds = data[1].sort((a, b) => a.title.localeCompare(b.title));
                let tags = data[2].sort((a, b) => a.tag.value.localeCompare(b.tag.value));

                if (this.popularity) {
                    this.popularityItems = tags.map(d => new Item(d.tag.id * -1, "/popular/tag/" + d.tag.id, d.tag.value));
                    this.popularityItems.concat(feeds.map(d => new Item(d.id, "/popular/feed/" + d.id, d.title)));
                }

                this.allItems = feeds.map(d => new Item(d.id, "/feed/" + d.id, d.title));

                if (tags.length > 0) {
                    let feedMap: Map<number, Feed> = new Map()
                    feeds.forEach((feed) => {
                        feedMap.set(feed.id, feed);
                    });

                    this.tags = tags.map(d =>
                         new Category(d.tag.id, d.tag.value, d.ids.map(id =>
                             new Item(id, "/tag/" + d.tag.id + "/feed/" + id, feedMap.get(id).title))));

                    this.tags.forEach(tag => this.collapses.set(tag.id, false));
                }
            },
            err => console.log(err),
        );
    }
}

class Category {
    constructor(public id: number, public title: string, public items: Item[]) {
    }
}

class Item {
    constructor(public id: number, public link: string, public title: string) {
    }
}