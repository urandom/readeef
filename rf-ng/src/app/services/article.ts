import { Injectable } from '@angular/core'
import { Response } from '@angular/http'
import { APIService, Serializable } from "./api";
import { Router, ActivatedRouteSnapshot, NavigationEnd, ParamMap, Data, Params } from '@angular/router';
import { Feed, FeedService } from "../services/feed"
import { EventService } from "../services/events"
import { QueryPreferences, PreferencesService } from "../services/preferences"
import { Observable, ConnectableObservable, Subject } from "rxjs";
import { listRoute, getArticleRoute } from "../main/routing-util"
import 'rxjs/add/observable/of'
import 'rxjs/add/operator/distinctUntilKeyChanged'
import 'rxjs/add/operator/combineLatest'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/scan'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/publishReplay'
import 'rxjs/add/operator/startWith'
import 'rxjs/add/operator/switchMap'

export class Article extends Serializable {
    id: number
    feed: string
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
    hits: Hits

    fromJSON(json) {
        if ("date" in json) {
            let date = json["date"];
            delete json["date"];

            this.date = new Date(date * 1000);

        }

        return super.fromJSON(json);
    }
}

interface Hits {
    fragments?: Fragments
}

interface Fragments {
    title?: Array<string>
    description?: Array<string>
}

class ArticlesResponse extends Serializable {
    articles: Article[]
}

class ArticleStateResponse extends Serializable {
    success: boolean
}

export interface QueryOptions {
    limit?: number,
    offset?: number,
    unreadFirst?: boolean,
    unreadOnly?: boolean,
    olderFirst?: boolean,
    afterID?: number,
    ids?: number[],
}

export interface Source {
    url : string
    updatable: boolean
}

export class UserSource {
    updatable = true
    get url() : string {
        return "";
    }
}

export class FavoriteSource {
    updatable = false
    get url() : string {
        return "/favorite";
    }
}

export class PopularSource {
    updatable = false
    constructor(private secondary: UserSource | FeedSource | TagSource) {}

    get url() : string {
        return "/popular" + this.secondary.url;
    }
}

export class FeedSource {
    updatable = true
    constructor(public readonly id : number) {}

    get url() : string {
        return `/feed/${this.id}`;
    }
}

export class TagSource {
    updatable = true
    constructor(public readonly id : number) {}

    get url() : string {
        return `/tag/${this.id}`;
    }
}

export class SearchSource {
    updatable = false
    constructor(private query: string, private secondary: UserSource | FeedSource | TagSource) {}

    get url() : string {
        return `/search${this.secondary.url}?query=${encodeURIComponent(this.query)}`;
    }
}

class ScanData {
    indexMap: Map<number, number> = new Map()
    articles: Array<Article> = []
}

interface ArticleProperty {
    id: number
    name: string
    value: any
}

interface ArticlesPayload {
    articles: Article[]
    fromEvent: boolean
}

@Injectable()
export class ArticleService {
    private articles : ConnectableObservable<Article[]>
    private paging = new Subject<number>()
    private stateChange = new Subject<ArticleProperty>()
    private page = 0
    private limit = 200
    private initialFetched = false

    constructor(
        private api: APIService,
        private feedService: FeedService,
        private eventService: EventService,
        private router: Router,
        private preferences: PreferencesService,
    ) {
        let queryPreferences = this.preferences.queryPreferences();
        let afterID = 0;

        let source = listRoute(this.router).map(
            route => this.nameToSource(route.data, route.params),
        ).filter(source =>
            source != null
        ).distinctUntilKeyChanged("url")

        this.articles = this.feedService.getFeeds().map(feeds =>
            feeds.reduce((map, feed) => {
                map[feed.id] = feed.title;

                return map;
            }, new Map<number, string>())
        ).switchMap(feedMap =>
            source.switchMap(source => {
                return queryPreferences.switchMap(prefs =>
                    Observable.merge(
                        this.paging.startWith(0).map(page => {
                            this.page = page
                            return page * this.limit
                        }).switchMap(offset => {
                            return this.getArticlesFor(source, prefs, this.limit, offset);
                        }).map(articles => <ArticlesPayload>{
                            articles: articles,
                            fromEvent: false,
                        }),
                        this.eventService.feedUpdate.filter(id =>
                            source.updatable
                        ).flatMap(id => 
                            this.getArticlesFor(new FeedSource(id), {
                                afterID: afterID,
                                olderFirst: prefs.olderFirst,
                                unreadOnly: prefs.unreadOnly,
                            }, this.limit, 0)
                        ).filter(
                            articles => articles.length > 0
                        ).map(articles => <ArticlesPayload>{
                            articles: articles,
                            fromEvent: true,
                        })
                    ).map(payload => {
                        payload.articles = payload.articles.map(article => {
                            if (article.hits && article.hits.fragments) {
                                if (article.hits.fragments.title.length > 0) {
                                    article.title = article.hits.fragments.title.join(" ")
                                }

                                if (article.hits.fragments.description.length > 0) {
                                    article.stripped = article.hits.fragments.description.join(" ")
                                }
                            }
                            if (!article.stripped) {
                                article.stripped = article.description.replace(/<[^>]+>/g, '');
                            }
                            article.feed = feedMap[article.feedID];

                            return article;
                        })

                        return payload;
                    }).scan((acc, payload) => {
                        if (payload.fromEvent) {
                            let updates = new Array<Article>();

                            for (let incoming of payload.articles) {
                                if (acc.indexMap.has(incoming.id)) {
                                    updates.push(incoming);
                                    break;
                                }

                                if (prefs.olderFirst) {
                                    for (let i = acc.articles.length - 1; i >= 0; i--) {
                                        if (this.shouldInsert(
                                            incoming, acc.articles[i], prefs
                                        )) {
                                            acc.articles.splice(i, 0, incoming)
                                            break
                                        }

                                    }
                                } else {
                                    for (let i = 0; i < acc.articles.length; i++) {

                                        if (this.shouldInsert(
                                            incoming, acc.articles[i], prefs
                                        )) {
                                            acc.articles.splice(i, 0, incoming)
                                            break
                                        }
                                    }
                                }
                            }

                            acc.indexMap = new Map()
                            for (let i = 0; i < acc.articles.length; i++) {
                                let article = acc.articles[i];
                                acc.indexMap[article.id] = i;
                                if (afterID < article.id) {
                                    afterID = article.id;
                                }
                            }

                            for (let update of updates) {
                                let idx = acc.indexMap[update.id];
                                acc.articles[idx] = update;
                            }
                        } else {
                            for (let article of payload.articles) {
                                if (acc.indexMap.has(article.id)) {
                                    let idx = acc.indexMap[article.id];
                                    acc.articles[idx] = article;
                                } else {
                                    acc.indexMap[article.id] = acc.articles.push(article) - 1;
                                }

                                if (afterID < article.id) {
                                    afterID = article.id;
                                }
                            }
                        }

                        return acc;
                    }, new ScanData()).combineLatest(
                        this.stateChange.startWith(null),
                        (data, propChange) => {
                            if (propChange != null) {
                                let idx = data.indexMap[propChange.id]
                                if (idx != -1) {
                                    data.articles[idx][propChange.name] = propChange.value;
                                }
                            }
                            return data;
                        }
                    ).map(data => data.articles).startWith([])
                )
            })
        ).publishReplay(1);

        this.articles.connect();
    }

    articleObservable() : Observable<Article[]> {
        return this.articles;
    }

    requestNextPage() {
        this.paging.next(this.page + 1);
    }

    public favor(id: number, favor: boolean) : Observable<Boolean> {
        return this.articleStateChange(id, "favorite", favor);
    }

    public read(id: number, read: boolean) {
        return this.articleStateChange(id, "read", read);
    }

    private articleStateChange(id: number, name: string, state: boolean) : Observable<Boolean> {
        let url = `article/${id}/${name}`
        let o : Observable<Response>
        if (state) {
            o = this.api.post(url);
        } else {
            o = this.api.delete(url);
        }

        return o.map(response =>
             new ArticleStateResponse().fromJSON(response.json()).success
        ).map(success => {
            if (success) {
                this.stateChange.next({
                    id: id,
                    name: name,
                    value: state,
                })
            }
            return success;
        });
    }

    private getArticlesFor(
        source: Source,
        prefs: QueryOptions,
        limit: number,
        offset: number,
    ): Observable<Article[]> {
        let options : QueryOptions = {
            limit: limit, offset: offset,
            unreadFirst: true,
            olderFirst: prefs.olderFirst,
            unreadOnly: prefs.unreadOnly,
            afterID: prefs.afterID,
        }

        if (source instanceof FavoriteSource) {
            options.unreadOnly = false
        }

        let res : Observable<Article[]>

        if (source instanceof PopularSource) {
            res = this.api.get(this.buildURL("article/popular" + source.url, options))
                .map(response => new ArticlesResponse().fromJSON(response.json()).articles);
        } else {
            res = this.api.get(this.buildURL("article" + source.url, options))
                .map(response => new ArticlesResponse().fromJSON(response.json()).articles);
        }

        if (!this.initialFetched) {
            this.initialFetched = true;

            let route = getArticleRoute([this.router.routerState.snapshot.root])

            if (route != null && +route.params["articleID"] > -1) {
                let id = +route.params["articleID"];

                return this.api.get(this.buildURL(
                    "article" + new UserSource().url, {ids: [id]}
                )).map(response => new ArticlesResponse()
                        .fromJSON(response.json()).articles[0]
                ).take(1).flatMap(initial => res.map(articles => {
                    if (!options.unreadOnly || !initial.read) {
                        for (let i = 0; i < articles.length; i++) {
                            let article = articles[i];
                            if (this.shouldInsert(initial, article, options)) {
                                articles.splice(i, 0, initial);
                                break;
                            }
                        }
                    }

                    articles.push(initial);
                    return articles;
                }));
            }
        }

        return res;
    }

    private buildURL(base: string, options?: QueryOptions) : string {
        if (!options) {
            options = {unreadFirst: true};
        }

        if (!options.limit) {
            options.limit = 200;
        }

        var query = new Array<string>();

        if (options.ids) {
            for (let id of options.ids) {
                query.push(`id=${id}`)
            }

            options.ids = undefined;
        }

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
            return base + (base.indexOf("?") == -1 ? "?" : "&") + query.join("&"); 
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
            case "search":
                return new SearchSource(decodeURIComponent(params["query"]), this.nameToSource(secondary, params));
            case "feed":
                return new FeedSource(params["id"]);
            case "tag":
                return new TagSource(params["id"]);
        }
    }

    private shouldInsert(incoming: Article, current: Article, options: QueryOptions) : boolean {
        if (options.unreadFirst && incoming.read != current.read) {
            return !incoming.read;
        }

        if (options.olderFirst) {
            if (incoming.date < current.date) {
                return true
            }
        } else {
            if (incoming.date > current.date) {
                return true
            }
        }

        return false
    }
}