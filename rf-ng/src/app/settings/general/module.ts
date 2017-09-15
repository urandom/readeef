import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { FormsModule }   from '@angular/forms';
import { RouterModule } from '@angular/router'
import { MdInputModule, MdButtonModule, MdSnackBarModule } from "@angular/material";
import { GeneralSettingsComponent, InvalidEmailSnack } from "./component"

@NgModule({
    declarations: [
        GeneralSettingsComponent,
        InvalidEmailSnack,
    ],
    entryComponents: [
        InvalidEmailSnack,
    ],
    imports: [
        CommonModule,
        FormsModule,
        RouterModule,
        MdInputModule,
        MdButtonModule,
        MdSnackBarModule,
    ],
    exports: [
        GeneralSettingsComponent,
        InvalidEmailSnack,
    ]
})
export class GeneralSettingsModule { }


