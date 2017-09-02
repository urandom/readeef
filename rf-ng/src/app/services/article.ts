import { Injectable } from '@angular/core'
import { APIService, Serializable } from "./api";
import { Router, ActivatedRouteSnapshot, NavigationEnd, ParamMap, Data, Params } from '@angular/router';
import { Feed, FeedService } from "../services/feed"
import { QueryPreferences, PreferencesService } from "../services/preferences"
import { Observable } from "rxjs";
import { BehaviorSubject } from "rxjs/BehaviorSubject";
import * as moment from 'moment';
import 'rxjs/add/observable/interval'
import 'rxjs/add/operator/combineLatest'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/scan'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/shareReplay'
import 'rxjs/add/operator/startWith'
import 'rxjs/add/operator/switchMap'

export class Article extends Serializable {
    id: number
    feedID: number
    title: string
    description: string
    stripped: string
    link: string
    date: Date
    read: boolean
    favorite: boolean
    thumbnail: string
    time: string

    fromJSON(json) {
        if ("date" in json) {
            let date = json["date"];
            delete json["date"];

            this.date = new Date(date * 1000);

        }

        return super.fromJSON(json);
    }
}

class ArticlesResponse extends Serializable {
    articles: Article[]
}

export interface QueryOptions {
    limit?: number,
    offset?: number,
    unreadFirst?: boolean,
    unreadOnly?: boolean,
    olderFirst?: boolean,
}

export interface Source {
    URL() : string
}

export class UserSource {
    URL() : string {
        return "";
    }
}

export class FavoriteSource {
    URL() : string {
        return "/favorite";
    }
}

export class PopularSource {
    constructor(private secondary: UserSource | FeedSource | TagSource) {}

    URL() : string {
        return "/popular" + this.secondary.URL();
    }
}

export class FeedSource {
    constructor(public readonly id : number) {}

    URL() : string {
        return `/feed/${this.id}`;
    }
}

export class TagSource {
    constructor(public readonly id : number) {}

    URL() : string {
        return `/tag/${this.id}`;
    }
}

class ScanData {
    indexMap: Map<number, number> = new Map()
    articles: Array<Article> = []
}

@Injectable()
export class ArticleService {
    private articles : Observable<Article[]>
    private paging: BehaviorSubject<number>
    private limit: number = 200

    constructor(
        private api: APIService,
        private feedService: FeedService,
        private router: Router,
        private preferences: PreferencesService,
    ) {
        let div = document.createElement('div');
        let queryPreferences = this.preferences.queryPreferences();
        this.paging = new BehaviorSubject(0);

        let source = this.router.events.filter(event =>
            event instanceof NavigationEnd
        ).map(e =>
            this.getLeafRouteData([this.router.routerState.snapshot.root])
        ).startWith(
            this.getLeafRouteData([this.router.routerState.snapshot.root])
        ).map(route => {
            return this.nameToSource(route.data, route.params)
        }).filter(source => source != null);

        this.articles = this.feedService.getFeeds().map(feeds =>
            feeds.reduce((map, feed) => {
                map[feed.id] = feed;

                return map;
            }, new Map<number, Feed>())
        ).switchMap(feedMap =>
            source.switchMap(source => {
                return queryPreferences.switchMap(prefs =>
                    this.paging.map(page =>
                        page * this.limit
                    ).switchMap(offset =>
                        this.getArticlesFor(source, prefs, this.limit, offset)
                        ).map(articles =>
                            articles.map(article => {
                                div.innerHTML = article.description;
                                article.time = moment(article.date).fromNow();
                                article.stripped = div.innerText;

                                return article;
                            })
                        ).scan((acc, articles) => {
                            for (let article of articles) {
                                if (acc.indexMap.has(article.id)) {
                                    let idx = acc.indexMap[article.id];
                                    acc.articles[idx] = article;
                                } else {
                                    acc.indexMap[article.id] = acc.articles.push(article) - 1;
                                }
                            }

                            return acc;
                        }, new ScanData()).map(data => data.articles)
                )
            })
        ).switchMap(articles =>
            Observable.interval(60000).startWith(0).map(v =>
                articles.map(article => {
                    article.time = moment(article.date).fromNow();
                    return article;
                })
            )
        ).shareReplay(1);
    }

    articleObservable() : Observable<Article[]> {
        return this.articles;
    }

    requestNextPage() {
        this.paging.next(this.paging.value + 1);
    }

    public favor(id: number, favor: boolean) {
        let url = `article/${id}/favorite`
        if (favor) {
            return this.api.post(url);
        }
        return this.api.delete(url);
    }

    public read(id: number, read: boolean) {
        let url = `article/${id}/read`
        if (read) {
            return this.api.post(url);
        }
        return this.api.delete(url);
    }

    private getArticlesFor(
        source: Source,
        prefs: QueryPreferences,
        limit: number,
        offset: number,
    ): Observable<Article[]> {
        let options : QueryOptions = {
            limit: limit, offset: offset,
            unreadFirst: true,
            olderFirst: prefs.olderFirst,
            unreadOnly: prefs.unreadOnly,
        }

        if (source instanceof PopularSource) {
            return this.api.get(this.buildURL("article/popular" + source.URL(), options))
                .map(response => new ArticlesResponse().fromJSON(response.json()).articles);
        } else {
            return this.api.get(this.buildURL("article" + source.URL(), options))
                .map(response => new ArticlesResponse().fromJSON(response.json()).articles);
        }
    }

    private buildURL(base: string, options?: QueryOptions) : string {
        if (!options) {
            options = {unreadFirst: true};
        }

        if (!options.limit) {
            options.limit = 200;
        }

        var query = new Array<string>();

        for (var i in options) {
            if (options.hasOwnProperty(i)) {
                let option = options[i];
                if (option === undefined) {
                    continue;
                }
                if (typeof option === "boolean") {
                    if (option) {
                        query.push(`${i}`);
                    }
                } else {
                    query.push(`${i}=${option}`);
                }
            }
        }

        if (query.length > 0) {
            return base + "?" + query.join("&"); 
        }

        return base;
    }

    private nameToSource(data: Data | string, params: Params) : Source {
        let name : string
        let secondary : string
        if (typeof data == "string") {
            name = data;
        } else {
            name = data["primary"];
            secondary = data["secondary"]
        }

        switch (name) {
            case "user":
                return new UserSource();
            case "favorite":
                return new FavoriteSource();
            case "popular":
                return new PopularSource(this.nameToSource(secondary, params));
            case "feed":
                return new FeedSource(params["id"]);
            case "tag":
                return new TagSource(params["id"]);
        }
    }

    private getLeafRouteData(routes: ActivatedRouteSnapshot[]) : ActivatedRouteSnapshot {
        for (let route of routes) {
            if ("primary" in route.data) {
                return route;
            }

            let r = this.getLeafRouteData(route.children);
            if (r != null) {
                return r;
            }
        }

        return null;
    }
}