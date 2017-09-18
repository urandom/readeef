import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { RouterModule } from '@angular/router'
import { MdSidenavModule, MdButtonModule, MdIconModule, MdToolbarModule } from "@angular/material";
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
import { routesModule } from "./routing";

@NgModule({
    declarations: [
        MainComponent,
    ],
    imports: [
        CommonModule,
        RouterModule,
        MdSidenavModule,
        MdButtonModule,
        MdIconModule,
        MdToolbarModule,
        NgbModule,
        SideBarModule,
        ToolbarModule,
        ArticleListModule,
        ArticleDisplayModule,
        SettingsModule,
        GeneralSettingsModule,
        DiscoverySettingsModule,
        ManagementSettingsModule,
        routesModule,
    ],
    exports: [
        MainComponent,
    ]
})
export class MainModule { }
