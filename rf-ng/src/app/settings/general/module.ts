import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { FormsModule, ReactiveFormsModule }   from '@angular/forms';
import {
    MatInputModule,
    MatButtonModule,
    MatDialogModule,
    MatSelectModule,
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
        MatInputModule,
        MatButtonModule,
        MatDialogModule,
        MatSelectModule,
    ],
    exports: [
        GeneralSettingsComponent,
        PasswordDialog,
    ]
})
export class GeneralSettingsModule { }


