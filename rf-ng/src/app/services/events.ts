/// <reference path="./eventsource.d.ts" />

import { Injectable } from '@angular/core'
import { Observable, Subject, BehaviorSubject } from "rxjs";
import { Serializable } from "./api";
import { TokenService } from './auth'
import 'rxjs/add/operator/combineLatest'
import 'rxjs/add/operator/startWith'

export class FeedUpdateEvent extends Serializable {
    feedID: number
    articleIDs: number[]
}

export class ArticleStateEvent extends Serializable {
    state: string
    value: boolean
    options: QueryOptions
}

export interface QueryOptions {
    ids?: number[]
    feedIDs?: number[]
    readOnly?: boolean
    unreadOnly?: boolean
    favoriteOnly?: boolean
    untaggedOnly?: boolean
    beforeID?: number
    afterID?: number
    beforeDate?: Date
    afterDate?: Date
}

@Injectable()
export class EventService {
    feedUpdate : Observable<FeedUpdateEvent>
    articleState : Observable<ArticleStateEvent>

    private eventSourceObservable : Observable<EventSource>
    private connectionSubject = new BehaviorSubject<boolean>(false);
    private refreshSubject = new Subject<any>();

    constructor(private tokenService : TokenService) {
        this.eventSourceObservable = this.tokenService.tokenObservable(
        ).combineLatest(
            this.refreshSubject.startWith(null), (token, v) => token,
        ).scan((source: EventSource, token :string) : EventSource => {
            if (source != null) {
                source.close()
            }

            if (token != "") {
                source = new EventSource("/api/v2/events?token=" + token)
                source['onopen'] = () => {
                    this.connectionSubject.next(true);
                };
                source['onerror'] = error => {
                    setTimeout(() => {
                        this.connectionSubject.next(false);
                        this.refresh();
                    }, 3000);
                };
            }

            return source
        }, <EventSource> null).filter(
            source => source != null
        ).shareReplay(1)

        this.feedUpdate = this.eventSourceObservable.flatMap(source => 
            Observable.fromEvent(source, "feed-update")
        ).map((event : MessageEvent) =>
            new FeedUpdateEvent().fromJSON(JSON.parse(event.data))
        )

        this.articleState = this.eventSourceObservable.flatMap(source => 
            Observable.fromEvent(source, "article-state-change")
        ).map((event: MessageEvent) =>
            new ArticleStateEvent().fromJSON(JSON.parse(event.data))
        )
    }

    connection() : Observable<boolean> {
        return this.connectionSubject.asObservable();
    }

    refresh() {
        this.refreshSubject.next(null);
    }
}