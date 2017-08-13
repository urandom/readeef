import { Component, OnInit } from "@angular/core";
import { Router, ActivatedRoute } from "@angular/router";
import { AlertService } from "../services/alert"
import { TokenService } from "../services/auth"

@Component({
    moduleId: module.id,
    templateUrl: "./login.html",
})
export class LoginComponent implements OnInit {
    user: string
    password: string
    loading = false

    returnURL: string

    constructor(
        private router: Router,
        private route: ActivatedRoute,
        private tokenService: TokenService,
        private alertService: AlertService,
    ) { }

    ngOnInit(): void {
        this.tokenService.delete();

        this.returnURL = this.route.snapshot.queryParams["returnURL"] || '/';
    }

    login() {
        this.loading = true
        this.tokenService.create(this.user, this.password).subscribe(
            data => this.router.navigate([this.returnURL]),
            error => {
                this.alertService.message(String(error), false);
                this.loading = false;
            }
        )
    }
}