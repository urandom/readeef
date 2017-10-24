import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import {
    MatButtonModule,
    MatDialogModule,
    MatIconModule,
    MatInputModule,
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
        MatButtonModule,
        MatDialogModule,
        MatIconModule,
        MatInputModule,
    ],
    exports: [
        ManagementSettingsComponent,
        ErrorDialog,
    ]
})
export class ManagementSettingsModule { }