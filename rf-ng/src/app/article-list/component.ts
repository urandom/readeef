import { Component, OnInit } from "@angular/core";
import { ActivatedRoute, ParamMap } from '@angular/router';
import { Article, ArticleService } from "../services/article"
import { Feed, FeedService } from "../services/feed"
import { ListItem } from './list-item';
import { ChangeEvent } from 'angular2-virtual-scroll';
import { Observable } from "rxjs";
import { BehaviorSubject } from "rxjs/BehaviorSubject";
import * as moment from 'moment';
import 'rxjs/add/observable/interval'
import 'rxjs/add/operator/scan'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/switchMap'

@Component({
    selector: "article-list",
    templateUrl: "./article-list.html",
    styleUrls: ["./article-list.css"],
})
export class ArticleListComponent implements OnInit {
    items: ListItem[] = []
    loading: boolean
    paging: BehaviorSubject<number>

    private limit: number = 200

    constructor(
        private articleService: ArticleService,
        private feedService: FeedService,
        private route: ActivatedRoute,
    ) {
        this.paging = new BehaviorSubject(0);
    }

    ngOnInit(): void {
        let div = document.createElement('div');

        this.feedService.getFeeds().map(feeds =>
            feeds.reduce((map, feed) => {
                map[feed.id] = feed;

                return map;
            }, new Map<number, Feed>())
        ).switchMap(feedMap =>
            this.route.data.switchMap(data => this.paging.map(page =>
                page * this.limit
            ).switchMap(offset =>
                this.getArticles(this.limit, offset)
            )).map(articles => 
                articles.map(article => {
                    div.innerHTML = article.description;
                    return <ListItem>{
                        title: article.title,
                        description: div.innerText,
                        thumbnail: article.thumbnail,
                        feed: feedMap[article.feedID].title,
                        date: article.date,
                        time: moment(article.date).fromNow(),
                        read: article.read,
                        favorite: article.favorite,
                    }
                })
            )
        ).subscribe(
            items => {
                this.loading = false;
                this.items = items;
            },
            error => {
                this.loading = false;
                console.log(error);
            }
        )
    }

    fetchMore(event: ChangeEvent) {
        if (event.end == this.items.length && !this.loading) {
            this.loading = true;
            this.paging.next(this.paging.value+1);
        }
    }

    private getArticles(limit: number, offset: number) : Observable<Article[]> {
        return this.articleService.getArticles({limit: limit, offset: offset});
    }
}