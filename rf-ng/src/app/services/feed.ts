import { Injectable } from '@angular/core'
import { Observable } from "rxjs";
import { APIService, Serializable } from "./api";
import 'rxjs/add/operator/map'

export class Feed {
    id: number
    title: string
    description: string
    link: string
    updateError: string
    subscribeError: string
}

class FeedsResponse extends Serializable {
    feeds: Feed[]
}

@Injectable()
export class FeedService {
    constructor(private api: APIService) { }

    getFeeds() : Observable<Feed[]> {
        return this.api.get("feed")
            .map(response => new FeedsResponse().fromJSON(response.json()).feeds);
    }
}