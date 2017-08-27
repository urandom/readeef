import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { RouterModule } from '@angular/router'
import { MdButtonModule, MdIconModule } from "@angular/material";
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { ArticleListComponent } from './component'
import { ListItemComponent } from './list-item'

@NgModule({
    declarations: [
        ArticleListComponent,
        ListItemComponent,
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
export class ArticleListModule { }
