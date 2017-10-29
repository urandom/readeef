import { Injectable } from '@angular/core'
import { Response } from '@angular/http'
import { APIService, Serializable } from "./api";
import { TokenService } from "./auth";
import { Router, ActivatedRouteSnapshot, NavigationEnd, ParamMap, Data, Params } from '@angular/router';
import { Feed, FeedService } from "../services/feed"
import { Tag, TagFeedIDs, TagService } from "../services/tag"
import { EventService, FeedUpdateEvent, QueryOptions as EventableQueryOptions } from "../services/events"
import { QueryPreferences, PreferencesService } from "../services/preferences"
import { Observable, BehaviorSubject, ConnectableObservable, Subject } from "rxjs";
import { listRoute, getArticleRoute } from "../main/routing-util"
import 'rxjs/add/observable/interval'
import 'rxjs/add/observable/of'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/distinctUntilChanged'
import 'rxjs/add/operator/distinctUntilKeyChanged'
import 'rxjs/add/operator/combineLatest'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/scan'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/publishReplay'
import 'rxjs/add/operator/shareReplay'
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
    format: ArticleFormat

    fromJSON(json) {
        if ("date" in json) {
            let date = json["date"];
            delete json["date"];

            this.date = new Date(date * 1000);

        }

        return super.fromJSON(json);
    }
}

export class ArticleFormat extends Serializable {
    keyPoints: string[]
    content: string
    topImage: string
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
    options: EventableQueryOptions,
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
    private paging = new BehaviorSubject<number>(0)
    private stateChange = new Subject<ArticleProperty>()
    private refresh = new BehaviorSubject<any>(null)
    private limit: number = 200
    private initialFetched = false
    private source : Observable<Source>

    constructor(
        private api: APIService,
        private tokenService: TokenService,
        private feedService: FeedService,
        private tagService: TagService,
        private eventService: EventService,
        private router: Router,
        private preferences: PreferencesService,
    ) {
        let queryPreferences = this.preferences.queryPreferences();

        this.source = listRoute(this.router).map(
            route => this.nameToSource(route.data, route.params),
        ).filter(source =>
            source != null
        ).distinctUntilKeyChanged("url");

        let feedsTagsObservable = this.tokenService.tokenObservable(
        ).switchMap(token =>
            this.feedService.getFeeds().combineLatest(
                this.tagService.getTagsFeedIDs(),
                (feeds, tags): [Map<number, string>, Map<number, number[]>] => {
                    let feedMap = feeds.reduce((map, feed) => {
                        map[feed.id] = feed.title;

                        return map;
                    }, new Map<number, string>());

                    let tagMap = tags.reduce((map, tagFeeds) => {
                        map[tagFeeds.tag.id] = tagFeeds.ids;

                        return map;
                    }, new Map<number, number[]>());

                    return [feedMap, tagMap]
                }
            )
        ).shareReplay(1);

        this.articles = feedsTagsObservable.switchMap(feedsTags =>
            this.source.combineLatest(
                this.refresh, (source, v) => source
            ).switchMap(source => {
                this.paging = new BehaviorSubject<number>(0);

                return queryPreferences.switchMap(prefs =>
                    Observable.merge(
                        this.paging.map(page => {
                            return page * this.limit;
                        }).switchMap(offset => {
                            return this.getArticlesFor(source, prefs, this.limit, offset);
                        }).map(articles => <ArticlesPayload>{
                            articles: articles,
                            fromEvent: false,
                        }),
                        this.eventService.feedUpdate.filter(event =>
                            this.shouldUpdate(event, source, feedsTags[1])
                        ).delay(30000).flatMap(event =>
                            this.getArticlesFor(new FeedSource(event.feedID), {
                                ids: event.articleIDs,
                                olderFirst: prefs.olderFirst,
                                unreadOnly: prefs.unreadOnly,
                            }, this.limit, 0)
                            ).map(articles => <ArticlesPayload>{
                                articles: articles,
                                fromEvent: true,
                            })
                    ).filter(p => p.articles != null).map(payload => {
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
                            article.feed = feedsTags[0][article.feedID];

                            return article;
                        })

                        return payload;
                    }).scan((acc, payload) => {
                        if (payload.fromEvent) {
                            let updates = new Array<Article>();

                            for (let incoming of payload.articles) {
                                if (acc.indexMap.has(incoming.id)) {
                                    updates.push(incoming);
                                    continue
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
                            }
                        }

                        return acc;
                    }, new ScanData()).combineLatest(
                        this.stateChange.startWith(null)
                            .distinctUntilChanged((a, b) => {
                                return JSON.stringify(a) == JSON.stringify(b);
                            }),
                        (data, propChange) => {
                            if (propChange != null) {
                                if (propChange.options.ids) {
                                    for (let id of propChange.options.ids) {
                                        let idx = data.indexMap[id]
                                        if (idx != undefined && idx != -1) {
                                            data.articles[idx][propChange.name] = propChange.value;
                                        }
                                    }
                                }

                                if (this.hasOptions(propChange.options)) {
                                    let tagged = new Set<number>()

                                    feedsTags[1].forEach(ids => {
                                        for (let id of ids) {
                                            tagged.add(id);
                                        }
                                    })

                                    for (let i = 0; i < data.articles.length; i++) {
                                        if (this.shouldSet(data.articles[i], propChange.options, tagged)) {
                                            data.articles[i][propChange.name] = propChange.value;
                                        }
                                    }
                                } else if (!propChange.options.ids) {
                                    for (let i = 0; i < data.articles.length; i++) {
                                        data.articles[i][propChange.name] = propChange.value;
                                    }
                                }
                            }
                            return data;
                        }
                        ).map(data => data.articles)
                ).startWith([])
            })
        ).publishReplay(1);

        this.articles.connect();

        var lastTick = new Date().getTime();
        Observable.interval(5000).subscribe(
            tick => {
                if (new Date().getTime() - lastTick > 10000) {
                    this.refreshArticles();
                }

                lastTick = new Date().getTime();
            }
        );

        this.eventService.articleState.subscribe(
            event => this.stateChange.next({
                    options: event.options,
                    name: event.state,
                    value: event.value,
            })
        )

        let lastNotification = new Date().getTime();
        Notification.requestPermission(p => {
            if (p == "granted") {
                feedsTagsObservable.combineLatest(
                    this.source,
                    (feedsTags, source) : [Map<number, string>, Map<number, number[]>, Source] => {
                        return [feedsTags[0], feedsTags[1], source];
                    }
                ).switchMap(data =>
                    this.eventService.feedUpdate.filter(event => 
                        this.shouldUpdate(event, data[2], data[1])
                    ).map(event => {
                        let title = data[0].get(event.feedID);
                        if (!title) {
                            return null;
                        }

                        return ["readeef: updates", `Feed ${title} has been updated`];
                    }).filter(msg => msg != null).delay(30000)
                ).subscribe(
                    msg => {
                        if (!document.hasFocus()) {
                            if (new Date().getTime() - lastNotification > 30000) {
                                new Notification(msg[0], {
                                    body: msg[1],
                                    icon: "/en/readeef.png",
                                    tag: "readeef",
                                })
                                lastNotification = new Date().getTime();
                            }
                        }
                    }
                );
            }
        });
    }

    articleObservable() : Observable<Article[]> {
        return this.articles;
    }

    requestNextPage() {
        this.paging.next(this.paging.value + 1);
    }

    formatArticle(id: number): Observable<ArticleFormat> {
        return this.api.get(`article/${id}/format`).map(
            response => new ArticleFormat().fromJSON(response.json())
        )
    }

    refreshArticles() {
        this.refresh.next(null)
    }

    public favor(id: number, favor: boolean) : Observable<Boolean> {
        return this.articleStateChange(id, "favorite", favor);
    }

    public read(id: number, read: boolean) {
        return this.articleStateChange(id, "read", read);
    }

    public readAll() {
        this.source.take(1).filter(
            source => source.updatable
        ).map(
            source => "article" + source.url + "/read"
        ).flatMap(
            url => this.api.post(url, JSON.stringify(true))
        ).map(response =>
            new ArticleStateResponse().fromJSON(response.json()).success
        ).subscribe(
            success => {},
            error => console.log(error),
        )
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
                    options: { ids: [id] },
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
            ids: prefs.ids,
        }

        let res = this.api.get(this.buildURL("article" + source.url, options))
            .map(response => new ArticlesResponse().fromJSON(response.json()).articles);

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
                    for (let i = 0; i < articles.length; i++) {
                        let article = articles[i];
                        if (article.id == initial.id) {
                            return articles
                        }
                        if (this.shouldInsert(initial, article, options)) {
                            articles.splice(i, 0, initial)
                            return articles
                        }
                    }

                    articles.push(initial)
                    return articles
                }));
            }
        }

        return res.catch(err => Observable.of(null));
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

    private shouldUpdate(event : FeedUpdateEvent, source: Source, tagMap: Map<number, number[]>) : boolean {
        let s = source
        if (s instanceof UserSource) {
            return true;
        } else if (source instanceof FeedSource) {
            if (event.feedID == source.id) {
                return true;
            }
        } else if (source instanceof TagSource) {
            let ids = tagMap[source.id];
            if (ids && ids.indexOf(event.feedID) != -1) {
                return true;
            }
        }

        return false;
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

    private hasOptions(options: EventableQueryOptions) : boolean {
        if (!options.feedIDs && !options.readOnly &&
            !options.unreadOnly && !options.favoriteOnly &&
            !options.untaggedOnly && !options.beforeID &&
            !options.afterID && !options.beforeDate && !options.afterDate) {
            return false;
        }

        return true;
    }

    private shouldSet(article: Article, options: EventableQueryOptions, tagged: Set<number>) : boolean {
        if (options.feedIDs && options.feedIDs.indexOf(article.feedID) == -1) {
            return false;
        }

        if (options.readOnly && !article.read) {
            return false;
        }

        if (options.unreadOnly && article.read) {
            return false;
        }

        if (options.favoriteOnly && !article.favorite) {
            return false;
        }

        if (options.untaggedOnly && !tagged.has(article.feedID)) {
            return false;
        }

        if (options.beforeID && article.id >= options.beforeID) {
            return false;
        }

        if (options.afterID && article.id <= options.afterID) {
            return false;
        }

        if (options.beforeDate && article.date >= options.beforeDate) {
            return false;
        }

        if (options.afterDate && article.date <= options.afterDate) {
            return false;
        }

        return true;
    }
}