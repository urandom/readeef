import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { RouterModule } from '@angular/router'
import { MdInputModule, MdButtonModule, MdIconModule, MdToolbarModule } from "@angular/material";
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { SideBarFeedComponent } from './feed-component'
import { SideBarSettingsComponent } from './settings-component'

@NgModule({
    declarations: [
        SideBarFeedComponent,
        SideBarSettingsComponent,
    ],
    imports: [
        CommonModule,
        RouterModule,
        MdInputModule,
        MdButtonModule,
        MdIconModule,
        MdToolbarModule,
        NgbModule,
    ],
    exports: [
        SideBarFeedComponent,
        SideBarSettingsComponent,
    ]
})
export class SideBarModule { }
