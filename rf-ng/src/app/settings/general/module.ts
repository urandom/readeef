import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { FormsModule, ReactiveFormsModule }   from '@angular/forms';
import {
    MdInputModule,
    MdButtonModule,
    MdDialogModule,
    MdSelectModule,
} from "@angular/material";
import {
    GeneralSettingsComponent,
    PasswordDialog,
} from "./component"

@NgModule({
    declarations: [
        GeneralSettingsComponent,
        PasswordDialog,

    ],
    entryComponents: [
        PasswordDialog,
    ],
    imports: [
        CommonModule,
        FormsModule,
        ReactiveFormsModule,
        MdInputModule,
        MdButtonModule,
        MdDialogModule,
        MdSelectModule,
    ],
    exports: [
        GeneralSettingsComponent,
        PasswordDialog,
    ]
})
export class GeneralSettingsModule { }


