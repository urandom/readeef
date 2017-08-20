import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { RouterModule } from '@angular/router'
import { MdInputModule, MdButtonModule, MdIconModule, MdToolbarModule } from "@angular/material";
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { SideBarComponent } from './component'

@NgModule({
    declarations: [
        SideBarComponent,
    ],
    imports: [
        CommonModule,
        RouterModule,
        MdInputModule,
        MdButtonModule,
        MdIconModule,
        MdToolbarModule,
        NgbModule,
    ],
    exports: [
        SideBarComponent,
    ]
})
export class SideBarModule { }
