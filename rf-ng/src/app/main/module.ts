import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { RouterModule } from '@angular/router'
import { MdSidenavModule, MdButtonModule, MdIconModule, MdToolbarModule } from "@angular/material";
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { MainComponent } from './component'
import { SideBarModule } from '../sidebar/module';

@NgModule({
    declarations: [
        MainComponent,
    ],
    imports: [
        CommonModule,
        RouterModule,
        MdSidenavModule,
        MdButtonModule,
        MdIconModule,
        MdToolbarModule,
        NgbModule,
        SideBarModule,
    ],
    exports: [
        MainComponent,
    ]
})
export class MainModule { }
