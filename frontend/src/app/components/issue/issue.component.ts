import { Component, ChangeDetectorRef, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';

@Component({
  selector: 'app-issue',
  standalone: true,
  imports: [FormsModule, CommonModule],
  template: `
    <div class="card">
      <h2>📤 Выдача товара</h2>

      <div class="section">
        <h3>Поиск товара</h3>
        <div class="form-row">
          <select [(ngModel)]="searchType" class="form-control">
            <option value="barcode">Штрихкод</option>
            <option value="storageCell">Ячейка</option>
            <option value="productBarcode">Баркод товара</option>
            <option value="article">Артикул</option>
          </select>
          <input [(ngModel)]="searchValue" class="form-control" placeholder="Введите значение..." (keyup.enter)="search()">
          <button class="btn-primary" (click)="search()" [disabled]="isSearching">
            {{ isSearching ? '⏳' : '🔍 Найти' }}
          </button>
        </div>
        <div *ngIf="searchError" class="error-msg">{{ searchError }}</div>
      </div>

      <div *ngIf="product" class="product-info">
        <h3>✅ Товар найден</h3>
        <div class="info-row"><span>Название:</span><b>{{ product.productName }}</b></div>
        <div class="info-row"><span>Ячейка:</span><b class="highlight">{{ product.cellLocation }}</b></div>
        <div class="info-row" *ngIf="product.barcode"><span>Штрихкод:</span><b>{{ product.barcode }}</b></div>
        <div class="info-row" *ngIf="product.articleNumber"><span>Артикул:</span><b>{{ product.articleNumber }}</b></div>

        <div *ngIf="product.issuedDate" class="already-issued">
          ⚠️ Товар уже выдан {{ product.issuedDate | date:'dd.MM.yyyy' }}
        </div>

        <div *ngIf="!product.issuedDate" class="issue-form">
          <h3>Оформить выдачу</h3>
          <div class="form-group">
            <label>Сотрудник:</label>
            <input [(ngModel)]="employeeName" class="form-control" placeholder="ФИО сотрудника">
          </div>
          <div class="form-group">
            <label>Вес при выдаче (кг):</label>
            <input [(ngModel)]="weightIssued" type="number" class="form-control" placeholder="Необязательно">
          </div>
          <div class="form-group">
            <label>Тип расходника:</label>
            <select [(ngModel)]="packageType" class="form-control">
              <option value="">— не выбран —</option>
              <optgroup label="Пакеты">
                <option value="пакет_S">Пакет S</option>
                <option value="пакет_M">Пакет M</option>
                <option value="пакет_L">Пакет L</option>
                <option value="пакет_XL">Пакет XL</option>
              </optgroup>
            </select>
            <div *ngIf="packageType && supplies.length > 0" class="stock-hint">
              <ng-container *ngFor="let s of supplies">
                <span *ngIf="s.packageType === packageType" [class.low]="s.count < 5">
                  На складе: <b>{{ s.count }} шт.</b>
                  <span *ngIf="s.count === 0"> ⚠️ Закончились!</span>
                  <span *ngIf="s.count > 0 && s.count < 5"> ⚠️ Мало</span>
                </span>
              </ng-container>
            </div>
          </div>
          <button class="btn-issue" (click)="issue()" [disabled]="!employeeName || isIssuing">
            {{ isIssuing ? '⏳ Оформляем...' : '📤 Выдать' }}
          </button>
          <div *ngIf="issueMessage" class="status-msg" [class.error]="isIssueError">{{ issueMessage }}</div>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .card { background: white; max-width: 600px; margin: 20px auto; padding: 25px; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.1); font-family: sans-serif; }
    h2 { color: #2c3e50; margin-bottom: 20px; }
    h3 { color: #475569; font-size: 15px; margin-bottom: 12px; }
    .section { margin-bottom: 20px; }
    .form-row { display: flex; gap: 8px; }
    .form-group { margin-bottom: 12px; }
    label { display: block; font-size: 13px; color: #64748b; margin-bottom: 4px; }
    .form-control { flex: 1; padding: 10px; border: 2px solid #e2e8f0; border-radius: 8px; font-size: 14px; box-sizing: border-box; width: 100%; }
    .btn-primary { padding: 10px 16px; background: #6366f1; color: white; border: none; border-radius: 8px; cursor: pointer; white-space: nowrap; font-weight: bold; }
    .btn-issue { width: 100%; padding: 12px; background: #f59e0b; color: white; border: none; border-radius: 8px; font-weight: bold; cursor: pointer; font-size: 15px; }
    .btn-issue:disabled { background: #cbd5e1; cursor: not-allowed; }
    .product-info { background: #f8fafc; border-radius: 8px; padding: 15px; border-left: 4px solid #22c55e; }
    .info-row { display: flex; justify-content: space-between; padding: 6px 0; border-bottom: 1px solid #e2e8f0; font-size: 14px; }
    .highlight { color: #6366f1; }
    .issue-form { margin-top: 15px; padding-top: 15px; border-top: 2px dashed #e2e8f0; }
    .already-issued { margin-top: 12px; padding: 10px; background: #fef9c3; border-radius: 6px; color: #92400e; font-weight: bold; }
    .error-msg { margin-top: 8px; color: #dc2626; font-size: 14px; }
    .status-msg { margin-top: 10px; text-align: center; font-weight: bold; color: #16a34a; }
    .status-msg.error { color: #dc2626; }
    .stock-hint { margin-top: 5px; font-size: 13px; color: #16a34a; }
    .stock-hint .low { color: #dc2626; }
  `]
})
export class IssueComponent implements OnInit {
  searchType = 'barcode';
  searchValue = '';
  isSearching = false;
  searchError = '';
  product: any = null;
  supplies: any[] = [];

  employeeName = '';
  weightIssued: number | null = null;
  packageType = '';
  isIssuing = false;
  issueMessage = '';
  isIssueError = false;

  constructor(private http: HttpClient, private cdr: ChangeDetectorRef) {}

  ngOnInit() {
    this.http.get<any[]>('/api/rashodniki').subscribe({
      next: (data) => { this.supplies = data; this.cdr.detectChanges(); },
      error: () => {}
    });
  }

  search() {
    if (!this.searchValue.trim()) return;
    this.isSearching = true;
    this.searchError = '';
    this.product = null;
    this.issueMessage = '';

    this.http.get<any>(`/api/issuance/search?searchType=${this.searchType}&searchValue=${encodeURIComponent(this.searchValue)}`)
      .subscribe({
        next: (data) => { this.product = data; this.isSearching = false; this.cdr.detectChanges(); },
        error: () => { this.searchError = '❌ Товар не найден'; this.isSearching = false; this.cdr.detectChanges(); }
      });
  }

  issue() {
    if (!this.employeeName || !this.product) return;
    this.isIssuing = true;
    this.issueMessage = '';
    const body: any = { productId: this.product.id, employeeName: this.employeeName };
    if (this.weightIssued) body.weightIssued = this.weightIssued;
    if (this.packageType) body.packageType = this.packageType;

    this.http.post<any>('/api/issuance/issue', body).subscribe({
      next: (res) => {
        this.issueMessage = `✅ ${res.message}`;
        this.isIssueError = false;
        this.product.issuedDate = new Date().toISOString();
        // Обновляем остатки локально
        if (this.packageType) {
          const s = this.supplies.find(x => x.packageType === this.packageType);
          if (s && s.count > 0) s.count--;
        }
        this.isIssuing = false;
        this.cdr.detectChanges();
      },
      error: (err) => {
        this.issueMessage = `❌ ${err.error?.message || 'Ошибка при выдаче'}`;
        this.isIssueError = true;
        this.isIssuing = false;
        this.cdr.detectChanges();
      }
    });
  }
}
