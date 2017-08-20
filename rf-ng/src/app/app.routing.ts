import { Routes, RouterModule } from '@angular/router'
import { environment } from '../environments/environment'

import { MainComponent } from './main/component';
import { LoginComponent } from './login/component';

import { AuthGuard } from './guards/auth';

export const AppRouting = RouterModule.forRoot([
    { path: "", component: MainComponent, canActivate: [AuthGuard] },
    { path: "login", component: LoginComponent }
], {enableTracing: !environment.production})