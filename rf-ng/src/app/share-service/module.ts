import { NgModule } from '@angular/core';
import { CommonModule }   from '@angular/common';
import { ShareServiceComponent } from "./component";

@NgModule({
    declarations: [
        ShareServiceComponent,
    ],
    imports: [
        CommonModule,
    ],
    exports: [
        ShareServiceComponent,
    ]
})
export class ShareServiceModule { }