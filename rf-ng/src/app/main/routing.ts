import { RouterModule, Routes, Route, Data } from '@angular/router';
import { SideBarFeedComponent } from "../sidebar/feed-component";
import { SideBarSettingsComponent } from "../sidebar/settings-component";
import { ToolbarFeedComponent } from "../toolbar/feed-component";
import { ToolbarSettingsComponent } from "../toolbar/settings-component";
import { ModuleWithProviders } from "@angular/core";
import { MainComponent } from "./component"
import { ArticleListComponent } from "../article-list/component"
import { ArticleDisplayComponent } from "../article-display/component"
import { SettingsComponent } from "../settings/component";
import { GeneralSettingsComponent } from "../settings/general/component";
import { DiscoverySettingsComponent } from "../settings/discovery/component";
import { ManagementSettingsComponent } from "../settings/management/component";

function createArtcleRoutes(paths: [string, Data][]) : Routes {
    let routes = []

    for (let [path, data] of paths) {
        routes.push({
            path: path,
            data: data,
            children: [
                {path: "", component: ArticleListComponent},
                {path: "article/:articleID", component: ArticleDisplayComponent},
            ],
        })
    }

    return routes;
}

export const routes: Routes = [
    {
        path: '',
        component: MainComponent,
        children: [
            {
                path: 'feed',
                children: [
                    {
                        path: "",
                        children: createArtcleRoutes([
                            ["", { "primary": "user" }],
                            ["search/:query", { "primary": "search", "secondary": "user" }],
                            ["favorite", { "primary": "favorite" }],
                            ["popular/tag/:id", { "primary": "popular", "secondary": "tag" }],
                            ["popular/:id", { "primary": "popular", "secondary": "feed" }],
                            ["popular", { "primary": "popular", "secondary": "user" }],
                            ["tag/:id", { "primary": "tag" }],
                            ["tag/:id/search/:query", { "primary": "search", "secondary": "tag" }],
                            [":id", { "primary": "feed" }],
                            [":id/search/:query", { "primary": "search", "secondary": "feed" }],
                        ])
                    },
                    {
                         path: "",
                         component: SideBarFeedComponent,
                         outlet: "sidebar" ,
                    },
                    {
                         path: "",
                         component: ToolbarFeedComponent,
                         outlet: "toolbar" ,
                    },
                ],
            },
            {
                path: 'settings',
                children: [
                    {
                        path: "",
                        component: SettingsComponent,
                        children: [
                            {
                                path: "general",
                                component: GeneralSettingsComponent,
                            },
                            {
                                path: "discovery",
                                component: DiscoverySettingsComponent,
                            },
                            {
                                path: "management",
                                component: ManagementSettingsComponent,
                            },
                            {
                                path: '',
                                redirectTo: 'general',
                                pathMatch: 'full',
                            },
                        ]
                    },
                    {
                         path: "",
                         component: SideBarSettingsComponent,
                         outlet: "sidebar"
                    },
                    {
                         path: "",
                         component: ToolbarSettingsComponent,
                         outlet: "toolbar"
                    },
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