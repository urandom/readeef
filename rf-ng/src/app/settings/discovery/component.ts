import { Component, ViewChild, ViewChildren, QueryList } from "@angular/core" ;
import { FormControl } from '@angular/forms';
import { MatCheckbox } from "@angular/material";
import { FeedService, Feed, OPMLimport, AddFeedResponse } from "../../services/feed";
import { Observable, Observer } from "rxjs";
import 'rxjs/add/operator/mergeMap'

@Component({
    selector: "settings-discovery",
    templateUrl: "./discovery.html",
    styleUrls: ["../common.css", "./discovery.css"]
})
export class DiscoverySettingsComponent {
    query = ""
    phase = "query"
    loading = false
    feeds = new Array<Feed>()
    emptySelection = false
    addFeedResult : AddFeedResponse

    queryFormControl = new FormControl('');

    @ViewChild("opmlInput")
    private opml;

    @ViewChildren(MatCheckbox)
    private feedChecks: QueryList<MatCheckbox>;

    constructor(
        private feedService: FeedService,
    ) {}

    search() {
        if (this.loading) {
            return;
        }

        if (this.query == "" && !this.opml.nativeElement.files.length) {
            this.queryFormControl.setErrors({"empty": true});
            return;
        }

        this.loading = true;

        let feedObservable : Observable<Feed[]>

        if (this.opml.nativeElement.files.length) {
            let file = this.opml.nativeElement.files[0];

            feedObservable = Observable.create((observer : Observer<OPMLimport>) => {
                let fileReader = new FileReader();

                fileReader.onload = function(event: Event) {
                    let contents = event.target["result"];

                    observer.next({opml: contents, dryRun: true});
                    observer.complete();
                }

                fileReader.onerror = function(event: ErrorEvent) : any {
                    observer.error(event);
                }

                fileReader.readAsText(file);
            }).flatMap(data => this.feedService.importOPML(data));
        } else {
            feedObservable = this.feedService.discover(this.query);
        }

        feedObservable.subscribe(
            feeds => {
                this.loading = false;
                this.phase = "search-result";
                this.feeds = feeds;
            },
            error => {
                this.loading = false;
                this.queryFormControl.setErrors({ "search": true });
                console.log(error);
            }
        )
    }

    add() {
        if (this.loading) {
            return;
        }

        let links = new Array<string>();
        this.feedChecks.forEach((check, index) => {
            if (check.checked) {
                links.push(this.feeds[index].link)
            }
        })

        if (links.length == 0) {
            this.emptySelection = true;
            return;
        }

        this.loading = true;

        this.feedService.addFeeds(links).subscribe(
            result => {
                this.loading = false;
                this.addFeedResult = result;
                this.phase = "add-result";
            },
            error => {
                this.loading = false;
                console.log(error);
            }
        )
    }

    baseURL(url: string) : string {
        let parser = document.createElement("a");
        parser.href = url;
        parser.pathname = "";

        return parser.href;
    }
}