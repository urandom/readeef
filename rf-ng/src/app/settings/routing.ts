import { RouterModule, Routes, Route } from '@angular/router';
import { ModuleWithProviders } from "@angular/core";
import { SettingsComponent } from "./component"

const routes: Routes = [
    {
        path: "general",
    },
    {
        path: '',
        redirectTo: 'general',
        pathMatch: 'full',
    },
]

export const routesModule : ModuleWithProviders = RouterModule.forChild(routes)