import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HttpClientModule } from '@angular/common/http';

import { AppComponent } from './components/app';
import { AppRouting } from './app.routing';

import { LoginComponent } from './login/component';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { MatInputModule, MatButtonModule, MatSidenavModule, MatIconModule, MatToolbarModule, MatDialogModule, MatSelectModule, MatCheckboxModule, MatSlideToggleModule, MatCardModule, MatMenuModule, MatProgressBarModule, MatSnackBarModule } from '@angular/material';
import { MainComponent } from './main/component';
import { CommonModule } from '@angular/common';
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { ArticleListComponent } from './article-list/component';
import { ArticleDisplayComponent } from './article-display/component';
import { SideBarFeedComponent } from './sidebar/feed-component';
import { ToolbarFeedComponent } from './toolbar/feed-component';
import { SettingsComponent } from './settings/component';
import { GeneralSettingsComponent, PasswordDialog } from './settings/general/component';
import { ManagementSettingsComponent, ErrorDialog } from './settings/management/component';
import { FiltersSettingsComponent, NewFilterDialog } from './settings/filters/component';
import { DiscoverySettingsComponent } from './settings/discovery/component';
import { ShareServiceComponent } from './share-service/component';
import { AdminSettingsComponent, NewUserDialog } from './settings/admin/component';
import { SideBarSettingsComponent } from './sidebar/settings-component';
import { ToolbarSettingsComponent } from './toolbar/settings-component';
import { ShareServicesSettingsComponent } from './settings/share-services/component';
import { ListItemComponent } from './article-list/list-item';
import { VirtualScrollModule } from 'angular2-virtual-scroll';

@NgModule({
  declarations: [
    AppComponent,
    LoginComponent,
    MainComponent,
    ArticleListComponent,
    ListItemComponent,
    ArticleDisplayComponent,
    SettingsComponent,
    GeneralSettingsComponent,
    PasswordDialog,
    DiscoverySettingsComponent,
    ManagementSettingsComponent,
    ErrorDialog,
    FiltersSettingsComponent,
    NewFilterDialog,
    ShareServicesSettingsComponent,
    ShareServiceComponent,
    AdminSettingsComponent,
    NewUserDialog,
    SideBarFeedComponent,
    SideBarSettingsComponent,
    ToolbarFeedComponent,
    ToolbarSettingsComponent,
  ],
  entryComponents: [
    ErrorDialog,
    NewFilterDialog,
    NewUserDialog,
    PasswordDialog,
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    HttpClientModule,
    AppRouting,
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    MatButtonModule,
    MatCardModule,
    MatCheckboxModule,
    MatDialogModule,
    MatIconModule,
    MatInputModule,
    MatMenuModule,
    MatProgressBarModule,
    MatSelectModule,
    MatSnackBarModule,
    MatSidenavModule,
    MatSlideToggleModule,
    MatToolbarModule,
    NgbModule.forRoot(),
    VirtualScrollModule,
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
