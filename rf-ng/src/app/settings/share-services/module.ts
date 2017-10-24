import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import {
    MatCardModule,
    MatSlideToggleModule,
} from "@angular/material";
import {
    ShareServicesSettingsComponent
} from "./component"

@NgModule({
    declarations: [
        ShareServicesSettingsComponent,
    ],
    imports: [
        CommonModule,
        MatCardModule,
        MatSlideToggleModule
    ],
    exports: [
        ShareServicesSettingsComponent,
    ]
})
export class ShareServicesSettingsModule { }


