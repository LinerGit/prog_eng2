import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SearchComponent } from './components/search/search.component';
import { LoginComponent } from './components/login/login.component';
import { IssueComponent } from './components/issue/issue.component';
import { ReturnComponent } from './components/return/return.component';
import { SuppliesComponent } from './components/supplies/supplies.component';
import { AuthService } from './services/auth.service';

@Component({
  selector: 'my-app',
  standalone: true,
  imports: [SearchComponent, LoginComponent, IssueComponent, ReturnComponent, SuppliesComponent, CommonModule],
  template: `
    <app-login *ngIf="!authService.currentUser" (loginSuccess)="onLogin()"></app-login>

    <div *ngIf="authService.currentUser">
      <div class="header">
        <span>Вы вошли как: <b>{{ authService.currentUser.username }}</b> ({{ authService.currentUser.role }})</span>
        <button (click)="logout()">Выйти</button>
      </div>
      <div class="tabs">
        <button [class.active]="activeTab === 'search'" (click)="activeTab = 'search'">🔍 Поиск</button>
        <button [class.active]="activeTab === 'issue'" (click)="activeTab = 'issue'">📤 Выдача</button>
        <button [class.active]="activeTab === 'return'" (click)="activeTab = 'return'">🔄 Возврат</button>
        <button [class.active]="activeTab === 'supplies'" (click)="activeTab = 'supplies'">🗂️ Расходники</button>
      </div>
      <app-search   *ngIf="activeTab === 'search'"></app-search>
      <app-issue    *ngIf="activeTab === 'issue'"></app-issue>
      <app-return   *ngIf="activeTab === 'return'"></app-return>
      <app-supplies *ngIf="activeTab === 'supplies'"></app-supplies>
    </div>
  `,
  styles: [`
    .header { background: #333; color: white; padding: 10px 20px; display: flex; justify-content: space-between; align-items: center; }
    .header button { background: #e74c3c; color: white; border: none; padding: 5px 10px; cursor: pointer; border-radius: 4px; }
    .tabs { display: flex; background: #f1f5f9; border-bottom: 2px solid #e2e8f0; padding: 0 20px; }
    .tabs button { padding: 12px 24px; border: none; background: transparent; cursor: pointer; font-size: 14px; font-weight: 600; color: #64748b; border-bottom: 3px solid transparent; margin-bottom: -2px; }
    .tabs button.active { color: #6366f1; border-bottom-color: #6366f1; }
  `]
})
export class AppComponent {
  activeTab = 'search';
  constructor(public authService: AuthService) {}
  onLogin() {}
  logout() { this.authService.logout(); }
}
