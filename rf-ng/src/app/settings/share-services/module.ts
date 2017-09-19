import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import {
    MdCardModule,
    MdSlideToggleModule,
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
        MdCardModule,
        MdSlideToggleModule
    ],
    exports: [
        ShareServicesSettingsComponent,
    ]
})
export class ShareServicesSettingsModule { }


