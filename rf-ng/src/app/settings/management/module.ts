import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import {
    MdButtonModule,
    MdIconModule,
    MdInputModule,
} from "@angular/material";
import {
    ManagementSettingsComponent
} from "./component"

@NgModule({
    declarations: [
        ManagementSettingsComponent,
    ],
    imports: [
        CommonModule,
        MdButtonModule,
        MdIconModule,
        MdInputModule,
    ],
    exports: [
        ManagementSettingsComponent,
    ]
})
export class ManagementSettingsModule { }