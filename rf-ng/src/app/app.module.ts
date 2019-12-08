import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HttpClientModule } from '@angular/common/http';

import { AppComponent } from './components/app';
import { AppRouting } from './app.routing';

import { LoginComponent } from './login/login.component';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { MatInputModule, MatButtonModule, MatSidenavModule, MatIconModule, MatToolbarModule, MatDialogModule, MatSelectModule, MatCheckboxModule, MatSlideToggleModule, MatCardModule, MatMenuModule, MatProgressBarModule, MatSnackBarModule } from '@angular/material';
import { MainComponent } from './main/main.component';
import { CommonModule } from '@angular/common';
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { ArticleListComponent } from './article-list/article-list.component';
import { ArticleDisplayComponent } from './article-display/article-display.component';
import { ToolbarFeedComponent } from './toolbar/toolbar.feed.component';
import { SettingsComponent } from './settings/settings.component';
import { GeneralSettingsComponent, PasswordDialog } from './settings/general/general.component';
import { ManagementSettingsComponent, ErrorDialog } from './settings/management/management.component';
import { FiltersSettingsComponent, NewFilterDialog } from './settings/filters/filters.component';
import { DiscoverySettingsComponent } from './settings/discovery/discovery.component';
import { ShareServiceComponent } from './share-service/share-service.component';
import { AdminSettingsComponent, NewUserDialog } from './settings/admin/admin.component';
import { SideBarSettingsComponent } from './sidebar/sidebar.settings.component';
import { ToolbarSettingsComponent } from './toolbar/toolbar.settings.component';
import { ShareServicesSettingsComponent } from './settings/share-services/share-services.component';
import { ListItemComponent } from './article-list/list-item.component';
import { VirtualScrollerModule } from 'ngx-virtual-scroller';
import { SideBarFeedComponent } from './sidebar/sidebar.feed.component';

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
    NgbModule,
    VirtualScrollerModule,
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
