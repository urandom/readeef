import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { FormsModule }   from '@angular/forms';
import { MdInputModule, MdSidenavModule, MdButtonModule, MdIconModule, MdToolbarModule } from "@angular/material";
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { HttpModule, BaseRequestOptions } from '@angular/http';

import { AppComponent } from './components/app';
import { AppRouting } from './app.routing';

import { HomeComponent } from './components/home';
import { LoginComponent } from './components/login';
import { NavMenuComponent } from './components/nav-menu';

import { AuthGuard } from './guards/auth';

import { TokenService } from './services/auth';
import { APIService } from './services/api';
import { FeaturesService } from './services/features';
import { FeedService } from './services/feed';
import { TagService } from './services/tag';

@NgModule({
  declarations: [
    AppComponent,
    HomeComponent,
    LoginComponent,
    NavMenuComponent,
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
    MdToolbarModule,
    NgbModule.forRoot(),
  ],
  providers: [
    TokenService,
    APIService,
    FeaturesService,
    FeedService,
    TagService,
    AuthGuard,
    BaseRequestOptions
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
