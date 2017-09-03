import { RouterModule, Routes, Route, Data } from '@angular/router';
import { SideBarFeedComponent } from "../sidebar/feed-component";
import { SideBarSettingsComponent } from "../sidebar/settings-component";
import { ToolbarFeedComponent } from "../toolbar/feed-component";
import { ToolbarSettingsComponent } from "../toolbar/settings-component";
import { ModuleWithProviders } from "@angular/core";
import { MainComponent } from "./component"
import { ArticleListComponent } from "../article-list/component"
import { ArticleDisplayComponent } from "../article-display/component"

const routes: Routes = [
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
                            ["favorite", { "primary": "favorite" }],
                            ["popular", { "primary": "popular", "secondary": "user" }],
                            ["popular/:id", { "primary": "popular", "secondary": "feed" }],
                            ["popular/tag/:id", { "primary": "popular", "secondary": "tag" }],
                            [":id", { "primary": "feed" }],
                            ["tag/:id", { "primary": "tag" }],
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

function createArtcleRoutes(paths: [string, Data][]) : Routes {
    let routes = []

    for (let path of paths) {
        routes.push({
            path: path[0],
            component: ArticleListComponent,
            data: path[1],
        });

        routes.push({
            path: path[0] + "article/:articleID",
            component: ArticleDisplayComponent,
            data: path[1],
        });
    }

    return routes;
}