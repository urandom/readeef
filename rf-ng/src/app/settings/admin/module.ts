import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { FormsModule, ReactiveFormsModule }   from '@angular/forms';
import {
    MatButtonModule,
    MatCheckboxModule,
    MatDialogModule,
    MatIconModule,
    MatInputModule,
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
        MatButtonModule,
        MatCheckboxModule,
        MatDialogModule,
        MatIconModule,
        MatInputModule,
    ],
    exports: [
        AdminSettingsComponent,
        NewUserDialog,
    ]
})
export class AdminSettingsModule { }