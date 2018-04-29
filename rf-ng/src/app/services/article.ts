import { Injectable } from '@angular/core'
import { Response } from '@angular/http'
import { APIService, Serializable } from "./api";
import { TokenService } from "./auth";
import { Router, ActivatedRouteSnapshot, NavigationEnd, ParamMap, Data, Params } from '@angular/router';
import { Feed, FeedService } from "../services/feed"
import { Tag, TagFeedIDs, TagService } from "../services/tag"
import { EventService, FeedUpdateEvent, QueryOptions as EventableQueryOptions } from "../services/events"
import { ListPreferences, PreferencesService } from "../services/preferences"
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
import { ObserveOnSubscriber } from 'rxjs/operators/observeOn';
import { fromEvent } from 'rxjs/observable/fromEvent';
import { QueryEncoder } from '@angular/http/src/url_search_params';

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
    time: string;
    hits: Hits;
    format: ArticleFormat;
	score?: number;
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

    fromJSON(json) {
        super.fromJSON(json);
        this.articles.map(article => {
            if (typeof article.date == "string") {
                article.date = new Date(article.date);
            }

            return article;
        });

        return this;
    }
}

interface Paging {
	time?: number;
	unreadTime?: number;
	score?: number;
	unreadScore?: number;
}

class ArticleStateResponse extends Serializable {
    success: boolean
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
    articles: Article[];
    fromEvent: boolean;
    unreadIDs?: Set<number>;
    unreadIDRange?: number[];
}

@Injectable()
export class ArticleService {
    private articles : ConnectableObservable<Article[]>
    private paging = new BehaviorSubject<any>(null)
    private stateChange = new Subject<ArticleProperty>()
    private updateSubject = new Subject<[QueryOptions, number]>()
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

        this.source = listRoute(this.router).filter(
            route => route != null
        ).map(
            route => this.nameToSource(route.data, route.params),
        ).filter(
            source => source != null
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
                        this.paging.flatMap(
                            v => this.datePaging(source, prefs.unreadFirst)
                        ).switchMap(paging => {
                            return this.getArticlesFor(source, {olderFirst: prefs.olderFirst, unreadOnly: prefs.unreadOnly}, this.limit, paging);
                        }).map(articles => <ArticlesPayload>{
                            articles: articles,
                            fromEvent: false,
                        }),
                        this.updateSubject.switchMap(opts =>
                             this.getArticlesFor(
                                 source, opts[0], this.limit, {}
                            ).combineLatest(
                                this.ids(source, { unreadOnly: true, beforeID: opts[0].afterID + 1, afterID: opts[1] - 1 }),
                                (articles, ids) => <ArticlesPayload>{
                                    articles: articles,
                                    unreadIDs: new Set(ids),
                                    unreadIDRange: [opts[1], opts[0].afterID],
                                    fromEvent: true,
                                }
                            )
                        ),
                        this.eventService.feedUpdate.filter(event =>
                            this.shouldUpdate(event, source, feedsTags[1])
                        ).delay(30000).flatMap(event =>
                            this.getArticlesFor(new FeedSource(event.feedID), {
                                ids: event.articleIDs,
                                olderFirst: prefs.olderFirst,
                                unreadOnly: prefs.unreadOnly,
                            }, this.limit, {})
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
                                if (article.id in acc.indexMap) {
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

        Observable.interval(10000).scan(
            (ticks, v) => {
                return [ticks[1], new Date().getTime()];
            }, [new Date().getTime(), new Date().getTime()]
        ).map(
            ticks => ticks[1] - ticks[0]
        ).switchMap(duration => {
            // Day
            if (duration > 86400000) {
                return this.eventService.connection().filter(
                    c => c
                ).first().map(c => null);
            // Minute
            } else if (duration > 60000) {
                return this.eventService.connection().filter(
                    c => c
                ).first().switchMap(v =>
                    this.articles.filter(
                        articles => articles.length > 0
                    ).map(
                        articles => articles.map(a => a.id)
                    ).map(ids =>
                        [Math.min.apply(Math, ids), Math.max.apply(Math, ids)]
                    ).first().map((minMax) : [QueryOptions, number] =>
                        this.preferences.olderFirst ?
                            [ { afterID: minMax[1] } , minMax[0]] :
                            [ { olderFirst: true, afterID: minMax[1] } , minMax[0]]
                    )
                );
            }

            return Observable.of(undefined);
        }).filter(
            opts => opts !== undefined
        ).subscribe(opts => {
            if (opts === null) {
                this.refreshArticles();
            }
            this.updateSubject.next(opts);
        });

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
        this.paging.next(null);
    }

    ids(source: Source, options: QueryOptions): Observable<number[]> {
        return this.api.get(this.buildURL(`article${source.url}/ids`, options))
            .map(response => {
                let ids : number[] = response.json()['ids'];
                return ids;
            })
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
            url => this.api.post(url)
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
        paging: Paging,
    ): Observable<Article[]> {
        let original : QueryOptions = Object.assign({}, prefs, {limit: limit});
        let options : QueryOptions = Object.assign({}, original);

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

        let res = this.api.get(this.buildURL("article" + source.url, options))
            .map(response => new ArticlesResponse().fromJSON(response.json()).articles);

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

            res = res.flatMap(articles => {
                if (articles.length == limit) {
                    return Observable.of(articles);
                }

                options.limit = limit - articles.length;

                return this.api.get(this.buildURL("article" + source.url, options))
                    .map(
                        response => new ArticlesResponse().fromJSON(response.json()).articles
                    ).map(
                        read => articles.concat(read)
                    )
            });
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
                    for (let i = 0; i < articles.length; i++) {
                        let article = articles[i];
                        if (article.id == initial.id) {
                            return articles
                        }
                        let prefs: ListPreferences = {
                            unreadFirst: !!paging.unreadTime,
                            unreadOnly: options.unreadOnly,
                            olderFirst: options.olderFirst,
                        };
                        if (this.shouldInsert(initial, article, prefs)) {
                            articles.splice(i, 0, initial)
                            return articles
                        }
                    }

                    articles.push(initial)
                    return articles
                }));
            }
        }

        return res.catch(err => {
            return Observable.of(null);
        });
    }

    private buildURL(base: string, options?: QueryOptions) : string {
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

    private datePaging(source: Source, unreadFirst: boolean) : Observable<Paging> {
        return this.articles.take(1).map(articles => {
            if (articles.length == 0) {
                // Initial query
				return unreadFirst ? {unreadTime: -1} : {};
            }

            let last = articles[articles.length - 1];
            let paging : Paging = {time: last.date.getTime() / 1000};

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
                    let prev = articles[i-1];
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
        });
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

    private shouldInsert(incoming: Article, current: Article, options: ListPreferences) : boolean {
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
