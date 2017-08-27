import { RouterModule, Routes } from '@angular/router';
import { SideBarFeedComponent } from "../sidebar/feed-component";
import { SideBarSettingsComponent } from "../sidebar/settings-component";
import { ModuleWithProviders } from "@angular/core";
import { MainComponent } from "./component"

const routes: Routes = [
    {
        path: '',
        component: MainComponent,
        children: [
            {
                path: 'feed',
                children: [
                    { path: "", component: SideBarFeedComponent, outlet: "sidebar" },
                ],
            },
            {
                path: 'settings',
                children: [
                    { path: "", component: SideBarSettingsComponent, outlet: "sidebar" },
                ],
            },
            {
                path: '',
                redirectTo: 'feed',
                pathMatch: 'full',
            },
        ],
    },
];

export const routesModule : ModuleWithProviders = RouterModule.forChild(routes);