import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { FormsModule, ReactiveFormsModule }   from '@angular/forms';
import {
    MatButtonModule,
    MatCheckboxModule,
    MatInputModule,
    MatProgressBarModule,
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
        MatButtonModule,
        MatCheckboxModule,
        MatInputModule,
        MatProgressBarModule,
    ],
    exports: [
        DiscoverySettingsComponent,
    ]
})
export class DiscoverySettingsModule { }


