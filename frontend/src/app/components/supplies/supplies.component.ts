import { Component, ChangeDetectorRef, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { AuthService } from '../../services/auth.service';

interface SupplyItem {
  packageType: string;
  label: string;
  group: string;
}

@Component({
  selector: 'app-supplies',
  standalone: true,
  imports: [FormsModule, CommonModule],
  template: `
    <div class="card">
      <h2>🗂️ Управление расходниками</h2>

      <div *ngIf="isLoading" class="loading">⏳ Загрузка...</div>

      <div *ngIf="!isLoading">
        <div *ngFor="let group of groups" class="group">
          <div class="group-title">{{ group }}</div>
          <div class="supply-row" *ngFor="let item of getByGroup(group)">
            <span class="supply-name">{{ item.label }}</span>
            <span class="count" [class.low]="getCount(item.packageType) < 5" [class.empty]="getCount(item.packageType) === 0">
              {{ getCount(item.packageType) }} шт.
            </span>
            <div class="controls">
              <input
                type="number"
                [(ngModel)]="addAmounts[item.packageType]"
                class="qty-input"
                placeholder="0"
                min="1">
              <button class="btn-add" (click)="addSupply(item.packageType)" [disabled]="!addAmounts[item.packageType] || addAmounts[item.packageType] < 1">+ Добавить</button>
              <button class="btn-remove" (click)="removeSupply(item.packageType)" [disabled]="!addAmounts[item.packageType] || addAmounts[item.packageType] < 1">− Убрать</button>
            </div>
          </div>
        </div>
      </div>

      <div *ngIf="message" class="msg" [class.error]="isError">{{ message }}</div>
    </div>
  `,
  styles: [`
    .card { background: white; max-width: 680px; margin: 20px auto; padding: 25px; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.1); font-family: sans-serif; }
    h2 { color: #2c3e50; margin-bottom: 20px; }
    .loading { color: #64748b; text-align: center; padding: 20px; }
    .group { margin-bottom: 20px; }
    .group-title { font-weight: 700; font-size: 13px; text-transform: uppercase; color: #94a3b8; letter-spacing: 1px; margin-bottom: 8px; padding-bottom: 4px; border-bottom: 1px solid #e2e8f0; }
    .supply-row { display: flex; align-items: center; gap: 10px; padding: 8px 0; border-bottom: 1px solid #f1f5f9; }
    .supply-name { flex: 1; font-size: 14px; color: #334155; }
    .count { min-width: 60px; text-align: center; font-weight: bold; font-size: 14px; color: #16a34a; }
    .count.low { color: #f59e0b; }
    .count.empty { color: #dc2626; }
    .controls { display: flex; gap: 6px; align-items: center; }
    .qty-input { width: 60px; padding: 6px 8px; border: 2px solid #e2e8f0; border-radius: 6px; font-size: 14px; text-align: center; }
    .btn-add { padding: 6px 12px; background: #10b981; color: white; border: none; border-radius: 6px; cursor: pointer; font-size: 13px; font-weight: bold; white-space: nowrap; }
    .btn-add:disabled { background: #cbd5e1; cursor: not-allowed; }
    .btn-remove { padding: 6px 12px; background: #ef4444; color: white; border: none; border-radius: 6px; cursor: pointer; font-size: 13px; font-weight: bold; white-space: nowrap; }
    .btn-remove:disabled { background: #cbd5e1; cursor: not-allowed; }
    .msg { margin-top: 15px; padding: 10px 15px; border-radius: 8px; background: #f0fdf4; color: #15803d; font-weight: bold; text-align: center; }
    .msg.error { background: #fef2f2; color: #b91c1c; }
  `]
})
export class SuppliesComponent implements OnInit {
  readonly allItems: SupplyItem[] = [
    { packageType: 'пакет_S',             label: 'Пакет S',             group: 'Пакеты' },
    { packageType: 'пакет_M',             label: 'Пакет M',             group: 'Пакеты' },
    { packageType: 'пакет_L',             label: 'Пакет L',             group: 'Пакеты' },
    { packageType: 'пакет_XL',            label: 'Пакет XL',            group: 'Пакеты' },
    { packageType: 'коробка',             label: 'Коробка',             group: 'Упаковка' },
    { packageType: 'сейф_пакет',          label: 'Сейф-пакет',          group: 'Упаковка' },
    { packageType: 'скотч',               label: 'Скотч',               group: 'Упаковка' },
    { packageType: 'нож',                 label: 'Нож',                 group: 'Инвентарь' },
    { packageType: 'ножницы',             label: 'Ножницы',             group: 'Инвентарь' },
    { packageType: 'тряпка',              label: 'Тряпка',              group: 'Инвентарь' },
    { packageType: 'средство_для_мытья',  label: 'Средство для мытья',  group: 'Инвентарь' },
  ];

  readonly groups = ['Пакеты', 'Упаковка', 'Инвентарь'];

  stockMap: Record<string, number> = {};
  addAmounts: Record<string, number> = {};
  isLoading = true;
  message = '';
  isError = false;

  constructor(
    private http: HttpClient,
    private cdr: ChangeDetectorRef,
    private authService: AuthService
  ) {}

  ngOnInit() { this.loadStock(); }

  loadStock() {
    this.isLoading = true;
    this.http.get<any[]>('/api/rashodniki').subscribe({
      next: (data) => {
        this.stockMap = {};
        data.forEach(r => this.stockMap[r.packageType] = r.count);
        this.isLoading = false;
        this.cdr.detectChanges();
      },
      error: () => { this.isLoading = false; this.cdr.detectChanges(); }
    });
  }

  private authHeaders(): { headers: HttpHeaders } {
    return { headers: new HttpHeaders({ 'Authorization': this.authService.getAuthHeader() }) };
  }

  getByGroup(group: string) {
    return this.allItems.filter(i => i.group === group);
  }

  getCount(packageType: string): number {
    return this.stockMap[packageType] ?? 0;
  }

  removeSupply(packageType: string) {
    const count = this.addAmounts[packageType];
    if (!count || count < 1) return;
    this.http.post<any>('/api/rashodniki/use', { packageType, count }, this.authHeaders()).subscribe({
      next: () => {
        this.stockMap[packageType] = Math.max(0, (this.stockMap[packageType] ?? 0) - count);
        this.addAmounts[packageType] = 0;
        const item = this.allItems.find(i => i.packageType === packageType);
        this.message = `✅ ${item?.label}: убрано ${count} шт. Итого: ${this.stockMap[packageType]}`;
        this.isError = false;
        this.cdr.detectChanges();
        setTimeout(() => { this.message = ''; this.cdr.detectChanges(); }, 3000);
      },
      error: (err) => { this.message = `❌ ${err.error?.message || 'Ошибка'}`; this.isError = true; this.cdr.detectChanges(); }
    });
  }

  addSupply(packageType: string) {
    const count = this.addAmounts[packageType];
    if (!count || count < 1) return;

    this.http.post<any>('/api/rashodniki/add', { packageType, count }, this.authHeaders()).subscribe({
      next: () => {
        this.stockMap[packageType] = (this.stockMap[packageType] ?? 0) + count;
        this.addAmounts[packageType] = 0;
        const item = this.allItems.find(i => i.packageType === packageType);
        this.message = `✅ ${item?.label}: добавлено ${count} шт. Итого: ${this.stockMap[packageType]}`;
        this.isError = false;
        this.cdr.detectChanges();
        setTimeout(() => { this.message = ''; this.cdr.detectChanges(); }, 3000);
      },
      error: (err) => {
        this.message = `❌ ${err.error?.message || 'Ошибка при добавлении'}`;
        this.isError = true;
        this.cdr.detectChanges();
      }
    });
  }
}
