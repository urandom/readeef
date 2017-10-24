import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { RouterModule } from '@angular/router'
import { MatButtonModule, MatIconModule } from "@angular/material";
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
        MatButtonModule,
        MatIconModule,
    ],
    exports: [
    ]
})
export class ArticleDisplayModule { }
