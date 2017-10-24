import { NgModule } from '@angular/core';
import { FormsModule }   from '@angular/forms';
import { CommonModule }   from '@angular/common';
import { RouterModule } from '@angular/router'
import {
    MatButtonModule,
    MatIconModule,
    MatToolbarModule,
    MatMenuModule,
    MatCheckboxModule,
    MatInputModule,
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
        MatButtonModule,
        MatIconModule,
        MatToolbarModule,
        MatMenuModule,
        MatCheckboxModule,
    ],
    exports: [
        ToolbarFeedComponent,
        ToolbarSettingsComponent,
    ]
})
export class ToolbarModule { }
