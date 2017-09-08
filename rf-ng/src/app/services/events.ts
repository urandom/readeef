/// <reference path="./eventsource.d.ts" />

import { Injectable } from '@angular/core'
import { Observable } from "rxjs";
import { TokenService } from './auth'

@Injectable()
export class EventService {
    feedUpdate : Observable<number>

    private eventSourceObservable : Observable<EventSource>

    constructor(private tokenService : TokenService) {
        this.eventSourceObservable = this.tokenService.tokenObservable(
        ).scan((source: EventSource, token :string) : EventSource => {
            if (source != null) {
                source.close()
            }

            if (token != "") {
                source = new EventSource("/api/v2/events?token=" + token)
            }

            return source
        }, <EventSource> null).filter(
            source => source != null
        ).shareReplay(1)

        this.feedUpdate = this.eventSourceObservable.flatMap(source => 
            Observable.fromEvent(source, "feed-update")
        ).map((event : DataEvent) => +event.data)
    }
}