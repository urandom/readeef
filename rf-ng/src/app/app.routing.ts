import { Routes, RouterModule } from '@angular/router'
import { environment } from '../environments/environment'

import { LoginComponent } from './login/login.component';

import { AuthGuard } from './guards/auth';
import { MainComponent } from './main/main.component';
import { ArticleListComponent } from './article-list/article-list.component';
import { ArticleDisplayComponent } from './article-display/article-display.component';
import { SideBarFeedComponent } from './sidebar/sidebar.feed.component';
import { ToolbarFeedComponent } from './toolbar/toolbar.feed.component';
import { SettingsComponent } from './settings/settings.component';
import { GeneralSettingsComponent } from './settings/general/general.component';
import { DiscoverySettingsComponent } from './settings/discovery/discovery.component';
import { ManagementSettingsComponent } from './settings/management/management.component';
import { FiltersSettingsComponent } from './settings/filters/filters.component';
import { ShareServicesSettingsComponent } from './settings/share-services/share-services.component';
import { AdminSettingsComponent } from './settings/admin/admin.component';
import { SideBarSettingsComponent } from './sidebar/sidebar.settings.component';
import { ToolbarSettingsComponent } from './toolbar/toolbar.settings.component';

export const AppRouting = RouterModule.forRoot([
    {
         path: "", canActivate: [AuthGuard],
        children: [{
            path: '',
            component: MainComponent,
            children: [
                {
                    path: 'feed',
                    children: [
                        {
                            path: "",
                            children: [
                                {
                                    path: "", data: { "primary": "user" }, children: [
                                        { path: "", component: ArticleListComponent },
                                        { path: "article/:articleID", component: ArticleDisplayComponent },
                                    ]
                                },
                                {
                                    path: "search/:query", data: { "primary": "search", "secondary": "user" }, children: [
                                        { path: "", component: ArticleListComponent },
                                        { path: "article/:articleID", component: ArticleDisplayComponent },
                                    ]
                                },
                                {
                                    path: "favorite", data: { "primary": "favorite" }, children: [
                                        { path: "", component: ArticleListComponent },
                                        { path: "article/:articleID", component: ArticleDisplayComponent },
                                    ]
                                },
                                {
                                    path: "popular/tag/:id", data: { "primary": "popular", "secondary": "tag" }, children: [
                                        { path: "", component: ArticleListComponent },
                                        { path: "article/:articleID", component: ArticleDisplayComponent },
                                    ]
                                },
                                {
                                    path: "popular/:id", data: { "primary": "popular", "secondary": "feed" }, children: [
                                        { path: "", component: ArticleListComponent },
                                        { path: "article/:articleID", component: ArticleDisplayComponent },
                                    ]
                                },
                                {
                                    path: "popular", data: { "primary": "popular", "secondary": "user" }, children: [
                                        { path: "", component: ArticleListComponent },
                                        { path: "article/:articleID", component: ArticleDisplayComponent },
                                    ]
                                },
                                {
                                    path: "tag/:id", data: { "primary": "tag" }, children: [
                                        { path: "", component: ArticleListComponent },
                                        { path: "article/:articleID", component: ArticleDisplayComponent },
                                    ]
                                },
                                {
                                    path: "tag/:id/search/:query", data: { "primary": "search", "secondary": "tag" }, children: [
                                        { path: "", component: ArticleListComponent },
                                        { path: "article/:articleID", component: ArticleDisplayComponent },
                                    ]
                                },
                                {
                                    path: ":id", data: { "primary": "feed" }, children: [
                                        { path: "", component: ArticleListComponent },
                                        { path: "article/:articleID", component: ArticleDisplayComponent },
                                    ]
                                },
                                {
                                    path: ":id/search/:query", data: { "primary": "search", "secondary": "feed" }, children: [
                                        { path: "", component: ArticleListComponent },
                                        { path: "article/:articleID", component: ArticleDisplayComponent },
                                    ]
                                },
                            ],
                        },
                        {
                            path: "",
                            component: SideBarFeedComponent,
                            outlet: "sidebar",
                        },
                        {
                            path: "",
                            component: ToolbarFeedComponent,
                            outlet: "toolbar",
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
                                    path: "filters",
                                    component: FiltersSettingsComponent,
                                },
                                {
                                    path: "share-services",
                                    component: ShareServicesSettingsComponent,
                                },
                                {
                                    path: "admin",
                                    component: AdminSettingsComponent,
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
    }

         ]
    },
    { path: "login", component: LoginComponent }
], {enableTracing: !environment.production})