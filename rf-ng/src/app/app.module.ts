import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { FormsModule }   from '@angular/forms';
import { MdInputModule, MdSidenavModule, MdButtonModule, MdIconModule } from "@angular/material";
import { HttpModule, BaseRequestOptions } from '@angular/http';

import { AppComponent } from './components/app';
import { AppRouting } from './app.routing';

import { AlertComponent } from './components/alert';
import { HomeComponent } from './components/home';
import { LoginComponent } from './components/login';

import { AuthGuard } from './guards/auth';

import { AlertService } from './services/alert';
import { TokenService } from './services/auth';

@NgModule({
  declarations: [
    AppComponent,
    AlertComponent,
    HomeComponent,
    LoginComponent,
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    FormsModule,
    HttpModule,
    AppRouting,
    MdInputModule,
    MdSidenavModule,
    MdButtonModule,
    MdIconModule,
  ],
  providers: [
    AlertService,
    TokenService,
    AuthGuard,
    BaseRequestOptions
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
