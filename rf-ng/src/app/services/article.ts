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
    olderFirst?: boolean,
}

@Injectable()
export class ArticleService {
    constructor(private api: APIService) { }

    public getArticles(options?: QueryOptions) : Observable<Article[]> {
        return this.api.get(this.buildURL("article", options))
            .map(response => new ArticlesResponse().fromJSON(response.json()).articles);
    }

    public getFavoriteArticles(options?: QueryOptions) : Observable<Article[]> {
        return this.api.get(this.buildURL("article/favorite", options))
            .map(response => new ArticlesResponse().fromJSON(response.json()).articles);
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
                if (typeof options[i] === "boolean") {
                    if (options[i]) {
                        query.push(`${i}`);
                    }
                } else {
                    query.push(`${i}=${options[i]}`);
                }
            }
        }

        if (query.length > 0) {
            return base + "?" + query.join("&"); 
        }

        return base;
    }
}