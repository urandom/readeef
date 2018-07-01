import { Injectable } from '@angular/core'
import { Observable } from "rxjs";
import { APIService } from "./api";
import { map } from 'rxjs/operators';


export class Feed {
    id: number;
    title: string;
    description: string;
    link: string;
    updateError: string;
    subscribeError: string;
}

export interface OPMLimport {
    opml: string;
    dryRun: boolean;
}

export interface AddFeedResponse {
    success: boolean;
    errors: AddFeedError[];
}

export class AddFeedError {
    link: string
    title: string
    error: string
}

interface FeedsResponse {
    feeds: Feed[]
}

interface OPMLResponse {
    opml: string;
}

interface AddFeedData {
    links: string[]
}

interface Status {
    success: boolean;
}

@Injectable({providedIn: "root"})
export class FeedService {
    constructor(private api: APIService) { }

    getFeeds() : Observable<Feed[]> {
        return this.api.get<FeedsResponse>("feed").pipe(
            map(response => response.feeds)
        );
    }

    discover(query: string) : Observable<Feed[]> {
        return this.api.get<FeedsResponse>(`feed/discover?query=${query}`).pipe(
            map(response => response.feeds)
        );
    }

    importOPML(data: OPMLimport): Observable<Feed[]> {
        var body = new FormData();
        body.append("opml", data.opml);
        if (data.dryRun) {
            body.append("dryRun", "true");
        }

        return this.api.post<FeedsResponse>("opml", body).pipe(
            map(response => response.feeds)
        );
    }

    exportOPML(): Observable<string> {
        return this.api.get<OPMLResponse>("opml").pipe(
            map(response => response.opml)
        );
    }

    addFeeds(links: string[]) : Observable<AddFeedResponse> {
        var body = new FormData();
        links.forEach(link => body.append("link", link));

        return this.api.post<AddFeedResponse>("feed", body);
    }

    deleteFeed(id: number) : Observable<boolean> {
        return this.api.delete<Status>(`feed/${id}`).pipe(
            map(response => response.success)
        );
    }

    updateTags(id: number, tags: string[]) : Observable<boolean> {
        var body = new FormData();
        tags.forEach(tag => body.append("tag", tag));

        return this.api.put<Status>(`feed/${id}/tags`, body).pipe(
            map(response => response.success)
        );
    }
}