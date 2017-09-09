import { NgModule } from '@angular/core';
import { FormsModule }   from '@angular/forms';
import { CommonModule }   from '@angular/common';
import { RouterModule } from '@angular/router'
import {
    MdButtonModule,
    MdIconModule,
    MdToolbarModule,
    MdMenuModule,
    MdCheckboxModule,
    MdInputModule,
} from "@angular/material";
import { ToolbarFeedComponent } from './feed-component'
import { ToolbarSettingsComponent } from './settings-component'

@NgModule({
    declarations: [
        ToolbarFeedComponent,
        ToolbarSettingsComponent,
    ],
    imports: [
        CommonModule,
        FormsModule,
        RouterModule,
        MdButtonModule,
        MdIconModule,
        MdToolbarModule,
        MdMenuModule,
        MdCheckboxModule,
    ],
    exports: [
        ToolbarFeedComponent,
        ToolbarSettingsComponent,
    ]
})
export class ToolbarModule { }
