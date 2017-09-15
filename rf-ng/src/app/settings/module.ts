import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { RouterModule } from '@angular/router'
import { SettingsComponent } from "./component"
//import { routesModule } from "./routing"

@NgModule({
    declarations: [
        SettingsComponent,
    ],
    imports: [
        CommonModule,
        RouterModule,
        //routesModule,
    ],
    exports: [
        SettingsComponent,
    ]
})
export class SettingsModule { }

