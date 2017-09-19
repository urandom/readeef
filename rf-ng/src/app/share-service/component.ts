import { Directive, OnInit, Input } from "@angular/core" 
import { SharingService, ShareService } from "../services/sharing"

@Directive({
    selector: "share-service",
})
export class ShareServiceComponent implements OnInit {
    @Input()
    id: string

    @Input()
    description: string

    @Input()
    category: string

    @Input()
    link: string

    @Input()
    share: string

    constructor(
        private sharingService: SharingService
    ) {}

    ngOnInit(): void {
        this.sharingService.register({
            id: this.id,
            description: this.description,
            category: this.category,
            link: this.link,
            template: this.share,
        })
    }
}
