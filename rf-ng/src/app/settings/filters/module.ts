import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { FormsModule, ReactiveFormsModule }   from '@angular/forms';
import {
    MatButtonModule,
    MatCheckboxModule,
    MatDialogModule,
    MatIconModule,
    MatInputModule,
    MatSelectModule,
    MatSlideToggleModule,
} from "@angular/material";
import {
    FiltersSettingsComponent,
    NewFilterDialog,
} from "./component"

@NgModule({
    declarations: [
        FiltersSettingsComponent,
        NewFilterDialog,
    ],
    entryComponents: [
        NewFilterDialog,
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
        MatSelectModule,
        MatSlideToggleModule,
    ],
    exports: [
        FiltersSettingsComponent,
        NewFilterDialog,
    ]
})
export class FiltersSettingsModule { }