import { NgModule } from '@angular/core';
import { FormsModule }   from '@angular/forms';
import { MatInputModule, MatButtonModule } from "@angular/material";
import { LoginComponent } from './component'

@NgModule({
  declarations: [
    LoginComponent,
  ],
  imports: [
    FormsModule,
    MatInputModule,
    MatButtonModule,
  ],
  exports: [
      LoginComponent,
  ]
})
export class LoginModule { }