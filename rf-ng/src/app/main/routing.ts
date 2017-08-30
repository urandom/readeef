import { RouterModule, Routes } from '@angular/router';
import { SideBarFeedComponent } from "../sidebar/feed-component";
import { SideBarSettingsComponent } from "../sidebar/settings-component";
import { ToolbarFeedComponent } from "../toolbar/feed-component";
import { ToolbarSettingsComponent } from "../toolbar/settings-component";
import { ModuleWithProviders } from "@angular/core";
import { MainComponent } from "./component"
import { ArticleListComponent } from "../article-list/component"

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
                         children: [
                              {path: "", component: ArticleListComponent, data: {"primary": "user"}},
                              {path: "favorite", component: ArticleListComponent, data: {"primary": "favorite"}},
                              {path: "popular", component: ArticleListComponent, data: {"primary": "popular", "secondary": "user"}},
                              {path: "popular/:id", component: ArticleListComponent, data: {"primary": "popular", "secondary": "feed"}},
                              {path: "popular/tag/:id", component: ArticleListComponent, data: {"primary": "popular", "secondary": "tag"}},
                              {path: ":id", component: ArticleListComponent, data: {"primary": "feed"}},
                              {path: "tag/:id", component: ArticleListComponent, data: {"primary": "tag"}},
                         ],
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