import { Injectable } from '@angular/core'
import { Observable } from "rxjs";
import { APIService, Serializable } from "./api";

export class Article extends Serializable {
    id: number
    feedID: number
    title: string
    description: string
    link: string
    date: Date
    read: boolean
    favorite: boolean
    thumbnail: string
    thumbnailLink: string

    fromJSON(json) {
        if ("date" in json) {
            let date = json["date"];
            delete json["date"];

            this.date = new Date(date * 1000);

        }

        return super.fromJSON(json);
    }
}

class ArticlesResponse extends Serializable {
    articles: Article[]
}

export interface QueryOptions {
    limit?: number,
    offset?: number,
    unreadFirst?: boolean,
    unreadOnly?: boolean,
    olderFirst?: boolean,
}

export interface Source {
    URL() : string
}

export class UserSource {
    URL() : string {
        return "";
    }
}

export class FavoriteSource {
    URL() : string {
        return "/favorite";
    }
}

export class PopularSource {
    constructor(private secondary: UserSource | FeedSource | TagSource) {}

    URL() : string {
        return "/popular" + this.secondary.URL();
    }
}

export class FeedSource {
    constructor(public readonly id : number) {}

    URL() : string {
        return `/feed/${this.id}`;
    }
}

export class TagSource {
    constructor(public readonly id : number) {}

    URL() : string {
        return `/tag/${this.id}`;
    }
}

@Injectable()
export class ArticleService {
    constructor(private api: APIService) { }

    public getArticles(source: UserSource | FavoriteSource | FeedSource | TagSource, options?: QueryOptions) : Observable<Article[]> {
        return this.api.get(this.buildURL("article" + source.URL(), options))
            .map(response => new ArticlesResponse().fromJSON(response.json()).articles);
    }

    public getPopularArticles(source: UserSource | FeedSource | TagSource, options?: QueryOptions) : Observable<Article[]> {
        return this.api.get(this.buildURL("article/popular" + source.URL(), options))
            .map(response => new ArticlesResponse().fromJSON(response.json()).articles);
    }

    public favor(id: number, favor: boolean) {
        let url = `article/${id}/favorite`
        if (favor) {
            return this.api.post(url);
        }
        return this.api.delete(url);
    }

    public read(id: number, read: boolean) {
        let url = `article/${id}/read`
        if (read) {
            return this.api.post(url);
        }
        return this.api.delete(url);
    }

    private buildURL(base: string, options?: QueryOptions) : string {
        if (!options) {
            options = {unreadFirst: true};
        }

        if (!options.limit) {
            options.limit = 200;
        }

        var query = new Array<string>();

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
            return base + "?" + query.join("&"); 
        }

        return base;
    }
}