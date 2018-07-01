import { Component, OnInit, OnDestroy } from '@angular/core';
import { UserService } from "../services/user"
import { Subscription } from "rxjs";
import { map } from 'rxjs/operators';

@Component({
    selector: 'side-bar',
    templateUrl: './side-bar-settings.html',
    styleUrls: ['./side-bar.css']
})
export class SideBarSettingsComponent implements OnInit, OnDestroy {
    admin: boolean

    private subscriptions = new Array<Subscription>()

    constructor(
        private userService: UserService,
    ) { }

    ngOnInit(): void {
        this.subscriptions.push(this.userService.getCurrentUser().pipe(map(
            user => user.admin
        )).subscribe(
            admin => this.admin = admin,
            error => console.log(error),
        ))
    }

    ngOnDestroy(): void {
        for (let subscription of this.subscriptions) {
            subscription.unsubscribe()
        }
    }
}