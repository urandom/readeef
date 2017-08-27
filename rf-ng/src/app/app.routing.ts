import { Routes, RouterModule } from '@angular/router'
import { environment } from '../environments/environment'

import { LoginComponent } from './login/component';

import { AuthGuard } from './guards/auth';

export const AppRouting = RouterModule.forRoot([
    {
         path: "", canActivate: [AuthGuard],
         loadChildren: './main/module#MainModule',
    },
    { path: "login", component: LoginComponent }
], {enableTracing: !environment.production})