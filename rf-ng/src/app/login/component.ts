import { Component, OnInit } from "@angular/core";
import { Router, ActivatedRoute } from "@angular/router";
import { TokenService } from "../services/auth"
import { HttpErrorResponse } from "@angular/common/http";

@Component({
    moduleId: module.id,
    templateUrl: "./login.html",
    styleUrls: ["./login.css"],
})
export class LoginComponent implements OnInit {
    user: string;
    password: string;
    loading = false;
    invalidLogin = false;

    returnURL: string;

    constructor(
        private router: Router,
        private route: ActivatedRoute,
        private tokenService: TokenService,
    ) { }

    ngOnInit(): void {
        this.tokenService.delete();

        this.returnURL = this.route.snapshot.queryParams["returnURL"] || '/';
    }

    login() {
        this.loading = true
        this.tokenService.create(this.user, this.password).subscribe(
            data => {
                this.invalidLogin = false;
                this.router.navigate([this.returnURL]);
            },
            (error: HttpErrorResponse) => {
                if (error.status == 401) {
                    this.invalidLogin = true;
                }
                this.loading = false;
            }
        )
    }
}