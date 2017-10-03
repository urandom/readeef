import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import {
    MdButtonModule,
    MdDialogModule,
    MdIconModule,
    MdInputModule,
} from "@angular/material";
import {
    ManagementSettingsComponent,
    ErrorDialog,
} from "./component"

@NgModule({
    declarations: [
        ManagementSettingsComponent,
        ErrorDialog,
    ],
    entryComponents: [
        ErrorDialog,
    ],
    imports: [
        CommonModule,
        MdButtonModule,
        MdDialogModule,
        MdIconModule,
        MdInputModule,
    ],
    exports: [
        ManagementSettingsComponent,
        ErrorDialog,
    ]
})
export class ManagementSettingsModule { }