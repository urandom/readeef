import { Injectable } from '@angular/core'
import { Router } from '@angular/router'
import { Observable } from "rxjs";
import { APIService } from "./api";
import { map } from 'rxjs/operators';


export class Tag {
    id: number;
    value: string;
}

interface TagsResponse {
    tags: Tag[];
}

export class TagFeedIDs {
    tag: Tag
    ids: number[]
}

interface TagsFeedIDs  {
    tagFeeds: TagFeedIDs[];
}

interface FeedIDsResponse  {
    feedIDs: number[];
}

@Injectable({providedIn: "root"})
export class TagService {
    constructor(private api: APIService) { }

    getTags() : Observable<Tag[]> {
        return this.api.get<TagsResponse>("tag").pipe(
            map(response => response.tags)
        );
    }

    getFeedIDs(tag: {id: number}): Observable<number[]> {
        return this.api.get<FeedIDsResponse>(`tag/${tag.id}/feedIDs`).pipe(
            map(response => response.feedIDs)
        );
    }

    getTagsFeedIDs(): Observable<TagFeedIDs[]> {
        return this.api.get<TagsFeedIDs>("tag/feedIDs").pipe(
            map(response => response.tagFeeds)
        );
    }
}