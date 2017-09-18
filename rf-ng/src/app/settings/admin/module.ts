import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { FormsModule, ReactiveFormsModule }   from '@angular/forms';
import {
    MdButtonModule,
    MdCheckboxModule,
    MdDialogModule,
    MdIconModule,
    MdInputModule,
} from "@angular/material";
import {
    AdminSettingsComponent,
    NewUserDialog,
} from "./component"

@NgModule({
    declarations: [
        AdminSettingsComponent,
        NewUserDialog,
    ],
    entryComponents: [
        NewUserDialog,
    ],
    imports: [
        CommonModule,
        FormsModule,
        ReactiveFormsModule,
        MdButtonModule,
        MdCheckboxModule,
        MdDialogModule,
        MdIconModule,
        MdInputModule,
    ],
    exports: [
        AdminSettingsComponent,
        NewUserDialog,
    ]
})
export class AdminSettingsModule { }