import { Injectable } from '@angular/core'
import { Router } from '@angular/router'
import { Observable } from "rxjs";
import { APIService, Serializable } from "./api";
import 'rxjs/add/operator/map'

export class Tag {
    id: number
    value: string
}

class TagsResponse extends Serializable {
    tags: Tag[]
}

export class TagFeedIDs {
    tag: Tag
    ids: number[]
}

class TagsFeedIDs extends Serializable {
    tagFeeds: TagFeedIDs[]
}

@Injectable()
export class TagService {
    constructor(private api: APIService) { }

    getTags() : Observable<Tag[]> {
        return this.api.get("tag")
            .map(response => new TagsResponse().fromJSON(response.json()).tags);
    }

    getTagsFeedIDs(): Observable<TagFeedIDs[]> {
        return this.api.get("tag/feedIDs")
            .map(response => new TagsFeedIDs().fromJSON(response.json()).tagFeeds)
    }
}