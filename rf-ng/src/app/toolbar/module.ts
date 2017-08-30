import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { RouterModule } from '@angular/router'
import { MdButtonModule, MdIconModule, MdToolbarModule, MdMenuModule } from "@angular/material";
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { ToolbarFeedComponent } from './feed-component'
import { ToolbarSettingsComponent } from './settings-component'

@NgModule({
    declarations: [
        ToolbarFeedComponent,
        ToolbarSettingsComponent,
    ],
    imports: [
        CommonModule,
        RouterModule,
        MdButtonModule,
        MdIconModule,
        MdToolbarModule,
        MdMenuModule,
    ],
    exports: [
        ToolbarFeedComponent,
        ToolbarSettingsComponent,
    ]
})
export class ToolbarModule { }
