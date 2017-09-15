import { Component, OnInit, OnDestroy } from "@angular/core" ;
import { MdSnackBar } from "@angular/material";
import { UserService } from "../../services/user";

@Component({
    selector: "settings-general",
    templateUrl: "./general.html",
    styleUrls: ["../common.css"]
})
export class GeneralSettingsComponent implements OnInit {
    firstName: string
    lastName: string
    email: string

    constructor(
        private userService: UserService,
        private snackBar: MdSnackBar,
    ) {}

    ngOnInit(): void {
        this.userService.getCurrentUser().subscribe(
            user => {
                this.firstName = user.firstName;
                this.lastName = user.lastName;
                this.email = user.email;
            },
            error => console.log(error),
        )
    }

    firstNameChange() {
        this.userService.setUserSetting(
            "first-name", this.firstName
        ).subscribe(
            success => {},
            error => console.log(error),
        )
    }

    lastNameChange() {
        this.userService.setUserSetting(
            "last-name", this.lastName
        ).subscribe(
            success => {},
            error => console.log(error),
        )
    }

    emailChange() {
        this.userService.setUserSetting(
            "email", this.email
        ).subscribe(
            success => {
                if (!success) {
                    this.snackBar.openFromComponent(InvalidEmailSnack, {duration: 2000})
                }
            },
            error => console.log(error),
        )
    }
}

@Component({
    templateUrl: "./invalid-email.html"
})
export class InvalidEmailSnack {
}