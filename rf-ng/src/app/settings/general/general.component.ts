import { Component, OnInit } from "@angular/core" ;
import { FormControl, Validators } from '@angular/forms';
import { MatDialog, MatDialogRef } from "@angular/material/dialog";
import { UserService } from "../../services/user";
import { HttpErrorResponse } from "@angular/common/http";

@Component({
    selector: "settings-general",
    templateUrl: "./general.html",
    styleUrls: ["../common.css", "./general.css"]
})
export class GeneralSettingsComponent implements OnInit {
    firstName: string
    lastName: string
    language: string

    emailFormControl = new FormControl('', [Validators.email]);

    constructor(
        private userService: UserService,
        private dialog: MatDialog,
    ) {
        this.language = localStorage.getItem("locale") || "en";
    }

    ngOnInit(): void {
        this.userService.getCurrentUser().subscribe(
            user => {
                this.firstName = user.firstName;
                this.lastName = user.lastName;
                this.emailFormControl.setValue(user.email);
            },
            error => console.log(error),
        )
    }

    firstNameChange() {
        this.userService.setUserSetting(
            "first-name", this.firstName
        ).subscribe(
            () => { },
            error => console.log(error),
        )
    }

    lastNameChange() {
        this.userService.setUserSetting(
            "last-name", this.lastName
        ).subscribe(
            () => { },
            error => console.log(error),
        )
    }

    emailChange() {
        this.emailFormControl.updateValueAndValidity()

        if (!this.emailFormControl.valid) {
            return
        }

        this.userService.setUserSetting(
            "email", this.emailFormControl.value
        ).subscribe(
            success => {
                if (!success) {
                    this.emailFormControl.setErrors({"email": true})
                }
            },
            error => console.log(error),
        )
    }

    languageChange(value: string) {
        location.href = "/" + value + "/";
    }

    changePassword() {
        this.dialog.open(PasswordDialog, {
            width: "250px",
        })
    }
}

@Component({
    templateUrl: "./password-form.html",
    styleUrls: ["../common.css", "./general.css"]
})
export class PasswordDialog {
    passwordConfirm: string

    currentFormControl = new FormControl('', [Validators.required]);
    passwordFormControl = new FormControl('', [Validators.required]);

    constructor(
        private dialogRef: MatDialogRef<PasswordDialog>,
        private userService: UserService,
    ) {}

    save() {
        if (this.passwordFormControl.value != this.passwordConfirm) {
            this.passwordFormControl.setErrors({"mismatch": true})
            return
        }

        this.currentFormControl.updateValueAndValidity()

        if (!this.currentFormControl.valid) {
            return
        }

        this.userService.changeUserPassword(
            this.passwordFormControl.value, this.currentFormControl.value
        ).subscribe(
            success => {
                if (!success) {
                    this.currentFormControl.setErrors({"auth": true});
                    return;
                }

                this.close();
            },
            (error: HttpErrorResponse) => {
                if (error.status == 400) {
                    this.currentFormControl.setErrors({"auth": true});
                    return;
                }
                console.log(error);
            },
        )
    }

    close() {
        this.dialogRef.close();
    }
}