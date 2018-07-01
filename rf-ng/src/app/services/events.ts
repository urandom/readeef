import { Injectable } from '@angular/core'
import { Observable, Subject, BehaviorSubject, fromEvent } from "rxjs";
import { TokenService } from './auth'
import { combineLatest, scan, filter, shareReplay, startWith, flatMap, map } from 'rxjs/operators';



export interface FeedUpdateEvent  {
    feedID: number;
    articleIDs: number[];
}

interface ArticleStateEvent  {
    state: string;
    value: boolean;
    options: QueryOptions;
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

@Injectable({providedIn: "root"})
export class EventService {
    feedUpdate : Observable<FeedUpdateEvent>
    articleState : Observable<ArticleStateEvent>

    private eventSourceObservable : Observable<EventSource>
    private connectionSubject = new BehaviorSubject<boolean>(false);
    private refreshSubject = new Subject<any>();

    constructor(private tokenService : TokenService) {
        this.eventSourceObservable = this.tokenService.tokenObservable().pipe(
            combineLatest(this.refreshSubject.pipe(startWith(null)), (token, v) => token),
            scan((source: EventSource, token: string): EventSource => {
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
            }, <EventSource> null),
            filter(source => source != null),
            shareReplay(1),
        );

        this.feedUpdate = this.eventSourceObservable.pipe(
            flatMap(source => fromEvent(source, "feed-update")),
            map((event : MessageEvent) => {
                let updateEvent : FeedUpdateEvent = JSON.parse(event.data);
                return updateEvent;
            })
        );

        this.articleState = this.eventSourceObservable.pipe(
            flatMap(source => fromEvent(source, "article-state-change")),
            map((event: MessageEvent) => {
                let stateEvent : ArticleStateEvent = JSON.parse(event.data);
                return stateEvent;
            })
        );
    }

    connection() : Observable<boolean> {
        return this.connectionSubject.asObservable();
    }

    refresh() {
        this.refreshSubject.next(null);
    }
}