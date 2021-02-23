import { Injectable } from '@angular/core';
import { Data, Params, Router } from '@angular/router';
import { BehaviorSubject, ConnectableObservable, merge, Observable, of, Subject } from "rxjs";
import { catchError, combineLatest, delay, distinctUntilChanged, distinctUntilKeyChanged, filter, first, mergeMap, map, publishReplay, scan, shareReplay, startWith, switchMap, take } from 'rxjs/operators';
import { getArticleRoute, listRoute } from "../main/routing-util";
import { EventService, FeedUpdateEvent, QueryOptions as EventableQueryOptions } from "../services/events";
import { FeedService } from "../services/feed";
import { ListPreferences, PreferencesService } from "../services/preferences";
import { TagService } from "../services/tag";
import { APIService } from "./api";
import { TokenService } from "./auth";

export class Article {
    id: number;
    feed: string;
    feedID: number;
    title: string;
    description: string;
    stripped: string;
    link: string;
    date: Date;
    read: boolean;
    favorite: boolean;
    thumbnail: string;
    thumbnailLink: string;
    time: string;
    hits: Hits;
    format: ArticleFormat;
    score?: number;
}

export interface ArticleFormat {
    keyPoints: string[];
    content: string;
    topImage: string;
}

interface Hits {
    fragments?: Fragments
}

interface Fragments {
    title?: Array<string>
    description?: Array<string>
}

interface ArticlesResponse {
    articles: Article[];
}

interface IDsResponse {
    ids: number[];
}

interface Paging {
    time?: number;
    unreadTime?: number;
    score?: number;
    unreadScore?: number;
}

interface ArticleStateResponse {
    success: boolean;
}

function processArticlesDates(articles: Article[]): Article[] {
    return articles.map(a => {
        if (typeof a.date == "string") {
            a.date = new Date(a.date);
        }
        return a;
    });
}

export interface QueryOptions {
    limit?: number;
    offset?: number;
    unreadOnly?: boolean;
    readOnly?: boolean;
    olderFirst?: boolean;
    ids?: number[];
    beforeID?: number
    afterID?: number;
    beforeTime?: number;
    afterTime?: number;
    beforeScore?: number;
    afterScore?: number;
}

export interface Source {
    url: string;
    updatable: boolean;
}

export class UserSource implements Source {
    updatable = true;
    get url(): string {
        return "";
    }
}

export class FavoriteSource implements Source {
    updatable = false
    get url(): string {
        return "/favorite";
    }
}

export class PopularSource implements Source {
    updatable = false
    constructor(private secondary: UserSource | FeedSource | TagSource) { }

    get url(): string {
        return "/popular" + this.secondary.url;
    }
}

export class FeedSource implements Source {
    updatable = true 
    constructor(public readonly id: number) { }

    get url(): string {
        return `/feed/${this.id}`;
    }
}

export class TagSource implements Source {
    updatable = true
    constructor(public readonly id: number) { }

    get url(): string {
        return `/tag/${this.id}`;
    }
}

export class SearchSource implements Source {
    updatable = false
    constructor(private query: string, private secondary: UserSource | FeedSource | TagSource) { }

    get url(): string {
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
    articles: Article[];
    fromEvent: boolean;
    unreadIDs?: Set<number>;
    unreadIDRange?: number[];
}

@Injectable({
    providedIn: "root",
})
export class ArticleService {
    private articles: ConnectableObservable<Article[]|true>
    private paging = new BehaviorSubject<any>(null)
    private stateChange = new Subject<ArticleProperty>()
    private updateSubject = new Subject<[QueryOptions, number]>()
    private refresh = new BehaviorSubject<any>(null)
    private limit: number = 200
    private initialFetched = false
    private source: Observable<Source>

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

        this.source = listRoute(this.router).pipe(
            filter(route => route != null),
            map(route => this.nameToSource(route.data, route.params)),
            filter(source => source != null),
            distinctUntilKeyChanged("url"),
        );

        let feedsTagsObservable = this.tokenService.tokenObservable().pipe(
            scan<string, [string, string, boolean]>((acc, token) => {
                var user = this.tokenService.tokenUser(token)
                acc[0] = token;
                acc[2] = acc[1] != user
                acc[1] = user;
                return acc
            }, ["", "", false]),
            filter(acc => acc[2]),
            switchMap(() =>
                this.feedService.getFeeds().pipe(combineLatest(
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
                ))
            ),
            shareReplay(1),
        )

        this.articles = feedsTagsObservable.pipe(
            switchMap(feedsTags =>
                this.source.pipe(
                    combineLatest(this.refresh, (source, _) => source),
                    switchMap(source => queryPreferences.pipe(
                            switchMap(prefs => {
                                this.paging = new BehaviorSubject<any>(0);

                                return merge(
                                    this.paging.pipe(
                                        mergeMap(_ => this.datePaging(source, prefs.unreadFirst)),
                                        switchMap(paging =>
                                            this.getArticlesFor(source, { olderFirst: prefs.olderFirst, unreadOnly: prefs.unreadOnly }, this.limit, paging)
                                        ),
                                        map(articles => <ArticlesPayload>{
                                            articles: articles,
                                            fromEvent: false,
                                        }),
                                    ),
                                    this.updateSubject.pipe(
                                        switchMap(opts =>
                                            this.getArticlesFor(
                                                source, opts[0], this.limit, {}
                                            ).pipe(combineLatest(
                                                this.ids(source, { unreadOnly: true, beforeID: opts[0].afterID + 1, afterID: opts[1] - 1, limit: 20000 }),
                                                (articles, ids) => <ArticlesPayload>{
                                                    articles: articles,
                                                    unreadIDs: new Set(ids),
                                                    unreadIDRange: [opts[1], opts[0].afterID],
                                                    fromEvent: true,
                                                }
                                            ))
                                        ),
                                    ),
                                    this.eventService.feedUpdate.pipe(
                                        filter(event => this.shouldUpdate(event, source, feedsTags[1])),
                                        delay(30000),
                                        mergeMap(event =>
                                            this.getArticlesFor(new FeedSource(event.feedID), {
                                                ids: event.articleIDs,
                                                olderFirst: prefs.olderFirst,
                                                unreadOnly: prefs.unreadOnly,
                                            }, this.limit, {})
                                        ),
                                        map(articles => <ArticlesPayload>{
                                            articles: articles,
                                            fromEvent: true,
                                        }),
                                    )
                                ).pipe(
                                    filter(p => p.articles != null),
                                    map(payload => {
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
                                    }),
                                    scan<ArticlesPayload, ScanData>((acc, payload) => {
                                        if (payload.unreadIDs) {
                                            for (let i = 0; i < acc.articles.length; i++) {
                                                let id = acc.articles[i].id;
                                                if (id >= payload.unreadIDRange[0] && id <= payload.unreadIDRange[1]) {
                                                    acc.articles[i].read = !payload.unreadIDs.has(id);
                                                }
                                            }
                                        }

                                        if (payload.fromEvent) {
                                            let updates = new Array<Article>();

                                            for (let incoming of payload.articles) {
                                                if (incoming.id in acc.indexMap) {
                                                    updates.push(incoming);
                                                    continue
                                                }

                                                if (prefs.olderFirst) {
                                                    for (let i = acc.articles.length - 1; i >= 0; i--) {
                                                        if (this.shouldInsert(
                                                            incoming, acc.articles[i], prefs
                                                        )) {
                                                            acc.articles.splice(i+1, 0, incoming)
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
                                                if (article.id in acc.indexMap) {
                                                    let idx = acc.indexMap[article.id];
                                                    acc.articles[idx] = article;
                                                } else {
                                                    acc.indexMap[article.id] = acc.articles.push(article) - 1;
                                                }
                                            }
                                        }

                                        return acc;
                                    }, new ScanData()),
                                    combineLatest(
                                        this.stateChange.pipe(
                                            startWith(null),
                                            distinctUntilChanged((a, b) => {
                                                return JSON.stringify(a) == JSON.stringify(b);
                                            })),
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
                                    ),
                                    map(data => data.articles),
                                    // Indicate that articles are being loaded
                                    startWith(true),
                                )
                            }),
                        )
                    )
                )
            ),
            publishReplay(1),
        ) as ConnectableObservable<Article[]>;

        this.articles.connect();

        this.eventService.connection().pipe(
            filter(c => c),
            switchMap(_ => this.articles.pipe(
                first(),
            )),
            filter(articles => articles !== true),
            map(articles => (articles as Article[])),
            filter(articles => articles.length > 0),
            map(articles => articles.map(a => a.id)),
            map(ids => [Math.min.apply(Math, ids), Math.max.apply(Math, ids)]),
            map((minMax): [QueryOptions, number] =>
                this.preferences.olderFirst ?
                    [{ afterID: minMax[1] }, minMax[0]] :
                    [{ olderFirst: true, afterID: minMax[1] }, minMax[0]]
            ),
        ).subscribe(
            opts => this.updateSubject.next(opts),
            err => console.log("Error refreshing article list after reconnect: ", err)
        )

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
                feedsTagsObservable.pipe(
                    combineLatest(
                        this.source,
                        (feedsTags, source): [Map<number, string>, Map<number, number[]>, Source] => {
                            return [feedsTags[0], feedsTags[1], source];
                        }
                    ),
                    switchMap(data =>
                        this.eventService.feedUpdate.pipe(
                            filter(event => this.shouldUpdate(event, data[2], data[1])),
                            map(event => {
                                let title = data[0].get(event.feedID);
                                if (!title) {
                                    return null;
                                }

                                return ["readeef: updates", `Feed ${title} has been updated`];
                            }),
                            filter(msg => msg != null),
                            delay(30000),
                        )
                    ),
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

    articleObservable(): Observable<Article[]|true> {
        return this.articles;
    }

    requestNextPage() {
        this.paging.next(null);
    }

    ids(source: Source, options: QueryOptions): Observable<number[]> {
        return this.api.get<IDsResponse>(
            this.buildURL(`article${source.url}/ids`, options)
        ).pipe(map(r => r.ids));
    }

    formatArticle(id: number): Observable<ArticleFormat> {
        return this.api.get<ArticleFormat>(`article/${id}/format`);
    }

    refreshArticles() {
        this.refresh.next(null)
    }

    public favor(id: number, favor: boolean): Observable<Boolean> {
        return this.articleStateChange(id, "favorite", favor);
    }

    public read(id: number, read: boolean) {
        return this.articleStateChange(id, "read", read);
    }

    public readAll() {
        this.source.pipe(
            take(1),
            filter(source => source.updatable),
            map(source => "article" + source.url + "/read"),
            mergeMap(url => this.api.post<ArticleStateResponse>(url)),
            map(response => response.success),
        ).subscribe(
            success => { },
            error => console.log(error),
        )
    }

    private articleStateChange(id: number, name: string, state: boolean): Observable<Boolean> {
        let url = `article/${id}/${name}`
        let o: Observable<ArticleStateResponse>
        if (state) {
            o = this.api.post<ArticleStateResponse>(url);
        } else {
            o = this.api.delete<ArticleStateResponse>(url);
        }

        return o.pipe(
            map(response => response.success),
            map(success => {
                if (success) {
                    this.stateChange.next({
                        options: { ids: [id] },
                        name: name,
                        value: state,
                    })
                }
                return success;
            }),
        );
    }

    private getArticlesFor(
        source: Source,
        prefs: QueryOptions,
        limit: number,
        paging: Paging,
    ): Observable<Article[]> {
        let original: QueryOptions = Object.assign({}, prefs, { limit: limit });
        let options: QueryOptions = Object.assign({}, original);

        let time = -1
        let score = 0;
        if (paging.unreadTime) {
            time = paging.unreadTime;
            score = paging.unreadScore;
            options.unreadOnly = true;
        } else if (paging.time) {
            time = paging.time;
            score = paging.score;
        }

        if (time != -1) {
            if (prefs.olderFirst) {
                options.afterTime = time;
                if (score) {
                    options.afterScore = score;
                }
            } else {
                options.beforeTime = time;
                if (score) {
                    options.beforeScore = score;
                }
            }
        }

        let res = this.api.get<ArticlesResponse>(this.buildURL("article" + source.url, options)).pipe(
            map(response => processArticlesDates(response.articles)),
        );

        if (paging.unreadTime) {
            options = Object.assign({}, original);
            options.readOnly = true;
            if (paging.time) {
                if (prefs.olderFirst) {
                    options.afterTime = paging.time;
                    if (paging.score) {
                        options.afterScore = paging.score;
                    }
                } else {
                    options.beforeTime = paging.time;
                    if (paging.score) {
                        options.beforeScore = paging.score;
                    }
                }
            }

            res = res.pipe(mergeMap(articles => {
                if (articles.length == limit) {
                    return of(articles);
                }

                options.limit = limit - articles.length;

                return this.api.get<ArticlesResponse>(this.buildURL("article" + source.url, options)).pipe(
                    map(response => processArticlesDates(response.articles)),
                    map(read => articles.concat(read))
                );
            }));
        }

        if (options.afterID) {
            res = res.pipe(mergeMap(articles => {
                if (!articles || !articles.length) {
                    return of(articles);
                }

                let maxID = Math.max.apply(Math, articles.map(a => a.id))

                let mod: QueryOptions = Object.assign({}, options, { afterID: maxID });
                return this.getArticlesFor(source, mod, limit, paging).pipe(
                    map(next => next && next.length ? next.concat(articles) : articles)
                );
            }))
        }

        if (!this.initialFetched) {
            this.initialFetched = true;

            let route = getArticleRoute([this.router.routerState.snapshot.root])

            if (route != null && +route.params["articleID"] > -1) {
                let id = +route.params["articleID"];

                return this.api.get<ArticlesResponse>(this.buildURL(
                    "article" + new UserSource().url, { ids: [id] }
                )).pipe(
                    map(response => processArticlesDates(response.articles)[0]),
                    take(1),
                    mergeMap(initial => res.pipe(
                        map(articles => {
                            if (options.unreadOnly) {
                                if (articles.findIndex(a => a.id == initial.id) != -1) {
                                    return;
                                }

                                if (options.olderFirst) {
                                    articles.unshift(initial);
                                } else {
                                    articles.push(initial);
                                }

                                return articles;
                            }

                            for (let i = 0; i < articles.length; i++) {
                                let article = articles[i];
                                if (article.id == initial.id) {
                                    return articles;
                                }
                                let prefs: ListPreferences = {
                                    unreadFirst: !!paging.unreadTime,
                                    unreadOnly: options.unreadOnly,
                                    olderFirst: options.olderFirst,
                                };
                                if (this.shouldInsert(initial, article, prefs)) {
                                    articles.splice(i, 0, initial)
                                    return articles;
                                }
                            }

                            articles.push(initial)
                            return articles;
                        }))
                    )
                );
            }
        }

        return res.pipe(catchError(err => {
            return of(null);
        }));
    }

    private buildURL(base: string, options?: QueryOptions): string {
        if (!options) {
            options = {};
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

    private nameToSource(data: Data | string, params: Params): Source {
        let name: string
        let secondary: string
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

    private datePaging(source: Source, unreadFirst: boolean): Observable<Paging> {
        return this.articles.pipe(
            take(1),
            map(articles => {
                if (articles === true || articles.length == 0) {
                    // Initial query
                    return unreadFirst ? { unreadTime: -1 } : {};
                }

                let last = articles[articles.length - 1];
                let paging: Paging = { time: last.date.getTime() / 1000 };

                if (source instanceof PopularSource) {
                    paging.score = last.score;
                }

                // The search indexes do not know which articles are read.
                if (unreadFirst && !(source instanceof SearchSource)) {
                    // fast-path
                    if (!last.read) {
                        paging.unreadTime = paging.time;
                        if (source instanceof PopularSource) {
                            paging.unreadScore = paging.score;
                        }
                        return paging;
                    }

                    for (let i = 1; i < articles.length; i++) {
                        let article = articles[i];
                        let prev = articles[i - 1];
                        if (article.read && !prev.read) {
                            paging.unreadTime = (prev.date.getTime() / 1000);
                            paging.unreadScore = prev.score;
                            break;
                        }
                    }

                    // no unread articles
                    if (!paging.unreadTime) {
                        paging.unreadTime = -1;
                    }
                }

                return paging;
            }),
        );
    }

    private hasOptions(options: EventableQueryOptions): boolean {
        if (!options.feedIDs && !options.readOnly &&
            !options.unreadOnly && !options.favoriteOnly &&
            !options.untaggedOnly && !options.beforeID &&
            !options.afterID && !options.beforeDate && !options.afterDate) {
            return false;
        }

        return true;
    }

    private shouldUpdate(event: FeedUpdateEvent, source: Source, tagMap: Map<number, number[]>): boolean {
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

    private shouldInsert(incoming: Article, current: Article, options: ListPreferences): boolean {
        if (options.unreadFirst && incoming.read != current.read) {
            return !incoming.read;
        }

        if (incoming.date > current.date) {
            return true
        }

        return false
    }

    private shouldSet(article: Article, options: EventableQueryOptions, tagged: Set<number>): boolean {
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
