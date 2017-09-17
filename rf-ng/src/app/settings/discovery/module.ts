import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { FormsModule, ReactiveFormsModule }   from '@angular/forms';
import { RouterModule } from '@angular/router'
import {
    MdButtonModule,
    MdCheckboxModule,
    MdInputModule,
    MdProgressBarModule,
} from "@angular/material";
import {
    DiscoverySettingsComponent
} from "./component"

@NgModule({
    declarations: [
        DiscoverySettingsComponent,
    ],
    imports: [
        CommonModule,
        FormsModule,
        ReactiveFormsModule,
        RouterModule,
        MdButtonModule,
        MdCheckboxModule,
        MdInputModule,
        MdProgressBarModule,
    ],
    exports: [
        DiscoverySettingsComponent,
    ]
})
export class DiscoverySettingsModule { }


