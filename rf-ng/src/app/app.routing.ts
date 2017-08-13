import { Routes, RouterModule } from '@angular/router'
import { environment } from '../environments/environment'

import { AlertComponent } from './components/alert';
import { HomeComponent } from './components/home';
import { LoginComponent } from './components/login';

import { AuthGuard } from './guards/auth';

export const AppRouting = RouterModule.forRoot([
    { path: "", component: HomeComponent, canActivate: [AuthGuard]},
    { path: "login", component: LoginComponent }
], {enableTracing: !environment.production})