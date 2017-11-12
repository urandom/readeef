import { Component, OnInit } from "@angular/core" ;
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { UserService, User } from "../../services/user";
import { MatDialog, MatDialogRef } from "@angular/material";
import { Observable, Subject } from "rxjs";
import 'rxjs/add/operator/combineLatest'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/startWith'

@Component({
    selector: "settings-admin",
    templateUrl: "./admin.html",
    styleUrls: ["../common.css", "./admin.css"]
})
export class AdminSettingsComponent implements OnInit {
    users = new Array<User>()
    current : User

    private refresher = new Subject<any>()

    constructor(
        private userService: UserService,
        private dialog: MatDialog,
    ) {}

    ngOnInit(): void {
        this.refresher.startWith(null).switchMap(
            v => this.userService.list()
        ).combineLatest(
            this.userService.getCurrentUser(),
            (users, current) =>
                users.filter(user => user.login != current.login)
        ).subscribe(
            users => this.users = users,
            error => console.log(error),
        );

        this.userService.getCurrentUser().subscribe(
            user => this.current = user,
            error => console.log(error),
        )
    }

    toggleActive(user: string, active: boolean) {
        this.userService.toggleActive(user, active).subscribe(
            success => {},
            error => console.log(error),
        )
    }

    deleteUser(event: Event, user: string) {
        // TODO: Replace with material dialog
        if (!confirm("Are you sure you want to delete user " + user)) {
            return;
        }

        this.userService.deleteUser(
            user
        ).subscribe(
            success => {
                if (success) {
                    let el = event.target["parentNode"];
                    while ((el = el.parentElement) && !el.classList.contains("user"));
                    el.parentNode.removeChild(el);
                }
            },
            error => console.log(error),
        );
    }

    newUser() {
        this.dialog.open(NewUserDialog, {
            width: "250px",
        }).afterClosed().subscribe(
            v => this.refresher.next(null),
        );
    }
}

@Component({
    templateUrl: "./new-user.html",
    styleUrls: ["../common.css", "./admin.css"]
})
export class NewUserDialog {
    form: FormGroup

    constructor(
        private dialogRef: MatDialogRef<NewUserDialog>,
        private userService: UserService,
        formBuilder: FormBuilder,
    ) {
        this.form = formBuilder.group({
            login: ['', Validators.required],
            password: ['', Validators.required],
        })
    }

    save() {
        if (!this.form.valid) {
            return;
        }

        let formModel = this.form.value;

        this.userService.addUser(
            formModel.login, formModel.password,
        ).subscribe(
            success => {
                if (success) {
                    this.close();
                }
            },
            error => console.log(error),
        )
    }

    close() {
        this.dialogRef.close();
    }
}