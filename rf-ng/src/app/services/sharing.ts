import { Injectable } from "@angular/core"
import { Article } from "./article"
import { Observable, BehaviorSubject } from "rxjs";

export interface ShareService {
    id: string
    description: string
    category: string
    link: string
    template: string
}

@Injectable()
export class SharingService {
    private services = new Map<string, ShareService>()
    private enabled : Set<string>
    private enabledSubject = new BehaviorSubject<ShareService[]>([])

    private static key = "enabled-services"

    constructor() {
        this.enabled = new Set<string>(JSON.parse(
            localStorage.getItem(SharingService.key)
        ));
    }

    register(service: ShareService) {
        this.services.set(service.id, service);

        this.enabledSubject.next(this.getEnabled());
    }

    enabledServices() : Observable<ShareService[]> {
        return this.enabledSubject.asObservable();
    }

    list() : [ShareService, boolean][] {
        return Array.from(this.services.values()).sort((a, b) => {
            let category = a.category.localeCompare(b.category);
            if (category == 0) {
                return a.id.localeCompare(b.id);
            }

            return category;
        }).map((service) : [ShareService, boolean] =>
            [service, this.isEnabled(service.id)]
        );
    }

    groupedList() : [ShareService, boolean][][] {
        let all = this.list();
        let groups = new Array<[ShareService, boolean][]>();
        let lastCategory = ""

        for (let item of all) {
            if (item[0].category != lastCategory) {
                lastCategory = item[0].category;
                groups.push(new Array<[ShareService, boolean]>());
            }

            let group = groups[groups.length - 1];
            group.push(item);
        }

        return groups;
    }

    toggle(id: string, enable: boolean) {
        if (this.services.has(id)) {
            if (enable) {
                this.enabled.add(id);
            } else {
                this.enabled.delete(id);
            }

            localStorage.setItem(
                SharingService.key,
                JSON.stringify(Array.from(this.enabled)),
            );

            this.enabledSubject.next(this.getEnabled());
        }
    }

    isEnabled(id: string): boolean {
        return this.enabled.has(id);
    }

    submit(id: string, article: Article) {
        let service = this.services.get(id);
        let url = service.template.replace(
            /{%\s*link\s*%}/g, article.link
        ).replace(
            /{%\s*title\s*%}/g, article.title
        )

        window.open(url, "_blank");
    }

    private getEnabled() : ShareService[] {
        return this.list().filter(s => s[1]).map(s => s[0]);
    }
}