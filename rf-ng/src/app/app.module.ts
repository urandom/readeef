import { BrowserModule, HAMMER_GESTURE_CONFIG, HammerModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HttpClientModule } from '@angular/common/http';

import { AppComponent } from './components/app';
import { AppRouting } from './app.routing';

import { LoginComponent } from './login/login.component';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatMenuModule } from '@angular/material/menu';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSelectModule } from '@angular/material/select';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { MatToolbarModule } from '@angular/material/toolbar';
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
import { GestureConfig } from "../gesture-config";

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
    HammerModule,
  ],
  bootstrap: [AppComponent],
  providers: [{ provide: HAMMER_GESTURE_CONFIG, useClass: GestureConfig }]
})
export class AppModule { }
