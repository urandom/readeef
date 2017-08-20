import { NgModule } from '@angular/core';
import { FormsModule }   from '@angular/forms';
import { MdInputModule, MdButtonModule } from "@angular/material";
import { LoginComponent } from './component'

@NgModule({
  declarations: [
    LoginComponent,
  ],
  imports: [
    FormsModule,
    MdInputModule,
    MdButtonModule,
  ],
  exports: [
      LoginComponent,
  ]
})
export class LoginModule { }