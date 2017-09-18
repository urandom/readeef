import { Injectable } from '@angular/core'

@Injectable()
export class FaviconService {
    private parser = document.createElement("a")

    iconURL(url: string) : string {
        this.parser.href = url;

        let domain = this.parser.hostname;

        return `//www.google.com/s2/favicons?domain=${domain}`
    }
}