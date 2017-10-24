import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { RouterModule } from '@angular/router'
import { MatSidenavModule, MatButtonModule, MatIconModule, MatToolbarModule } from "@angular/material";
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { MainComponent } from './component'
import { SideBarModule } from '../sidebar/module';
import { ToolbarModule } from '../toolbar/module';
import { ArticleListModule } from '../article-list/module';
import { ArticleDisplayModule } from '../article-display/module';
import { SettingsModule } from "../settings/module";
import { GeneralSettingsModule } from "../settings/general/module";
import { DiscoverySettingsModule } from "../settings/discovery/module";
import { ManagementSettingsModule } from "../settings/management/module";
import { FiltersSettingsModule } from "../settings/filters/module";
import { ShareServicesSettingsModule } from "../settings/share-services/module";
import { AdminSettingsModule } from "../settings/admin/module";
import { ShareServiceModule } from "../share-service/module";
import { routesModule } from "./routing";

@NgModule({
    declarations: [
        MainComponent,
    ],
    imports: [
        CommonModule,
        RouterModule,
        MatSidenavModule,
        MatButtonModule,
        MatIconModule,
        MatToolbarModule,
        NgbModule,
        SideBarModule,
        ToolbarModule,
        ArticleListModule,
        ArticleDisplayModule,
        SettingsModule,
        GeneralSettingsModule,
        DiscoverySettingsModule,
        ManagementSettingsModule,
        FiltersSettingsModule,
        ShareServicesSettingsModule,
        AdminSettingsModule,
        ShareServiceModule,
        routesModule,
    ],
    exports: [
        MainComponent,
    ]
})
export class MainModule { }
