import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { RouterModule } from '@angular/router'
import { MdButtonModule, MdIconModule } from "@angular/material";
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { ArticleDisplayComponent } from './component'

@NgModule({
    declarations: [
        ArticleDisplayComponent,
    ],
    imports: [
        CommonModule,
        RouterModule,
        NgbModule,
        MdButtonModule,
        MdIconModule,
    ],
    exports: [
    ]
})
export class ArticleDisplayModule { }
