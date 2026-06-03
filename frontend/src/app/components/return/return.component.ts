import { Component, ChangeDetectorRef } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';

@Component({
  selector: 'app-return',
  standalone: true,
  imports: [FormsModule, CommonModule],
  template: `
    <div class="card">
      <h2>🔄 Возврат товара</h2>
      <p class="hint">Срок возврата — <b>14 дней</b> с даты выдачи</p>

      <div class="form-group">
        <label>Штрихкод:</label>
        <input [(ngModel)]="barcode" class="form-control" placeholder="Штрихкод товара">
      </div>
      <div class="form-group">
        <label>Артикул:</label>
        <input [(ngModel)]="articleNumber" class="form-control" placeholder="Необязательно">
      </div>
      <div class="form-group">
        <label>Ячейка:</label>
        <input [(ngModel)]="cellLocation" class="form-control" placeholder="Например A-01">
      </div>
      <div class="form-group">
        <label>Дата выдачи:</label>
        <input [(ngModel)]="issuedDate" type="date" class="form-control">
      </div>

      <button class="btn-return" (click)="processReturn()" [disabled]="!isFormValid() || isLoading">
        {{ isLoading ? '⏳ Оформляем...' : '🔄 Оформить возврат' }}
      </button>

      <div *ngIf="resultMessage" class="result-msg" [class.accepted]="resultStatus === 'accepted'" [class.rejected]="resultStatus === 'rejected'">
        <b>{{ resultStatus === 'accepted' ? '✅' : '❌' }}</b> {{ resultMessage }}
        <div *ngIf="daysLeft !== null && resultStatus === 'accepted'" class="days-left">
          Осталось дней на возврат: <b>{{ daysLeft }}</b>
        </div>
      </div>

      <div *ngIf="history.length > 0" class="history">
        <h3>История возвратов</h3>
        <div *ngFor="let r of history" class="history-row" [class.accepted]="r.status === 'accepted'" [class.rejected]="r.status === 'rejected'">
          <span>{{ r.barcode }}</span>
          <span>{{ r.returnDate | date:'dd.MM.yyyy HH:mm' }}</span>
          <span class="badge">{{ r.status === 'accepted' ? 'Принят' : 'Отклонён' }}</span>
        </div>
      </div>

      <button *ngIf="barcode" class="btn-history" (click)="loadHistory()">📋 История по штрихкоду</button>
    </div>
  `,
  styles: [`
    .card { background: white; max-width: 600px; margin: 20px auto; padding: 25px; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.1); font-family: sans-serif; }
    h2 { color: #2c3e50; margin-bottom: 5px; }
    .hint { color: #64748b; font-size: 13px; margin-bottom: 20px; }
    .form-group { margin-bottom: 12px; }
    label { display: block; font-size: 13px; color: #64748b; margin-bottom: 4px; }
    .form-control { width: 100%; padding: 10px; border: 2px solid #e2e8f0; border-radius: 8px; font-size: 14px; box-sizing: border-box; }
    .btn-return { width: 100%; padding: 12px; background: #10b981; color: white; border: none; border-radius: 8px; font-weight: bold; cursor: pointer; font-size: 15px; margin-top: 5px; }
    .btn-return:disabled { background: #cbd5e1; cursor: not-allowed; }
    .btn-history { margin-top: 15px; width: 100%; padding: 8px; background: transparent; border: 2px solid #6366f1; color: #6366f1; border-radius: 8px; cursor: pointer; font-weight: bold; }
    .result-msg { margin-top: 15px; padding: 12px 15px; border-radius: 8px; font-size: 14px; }
    .result-msg.accepted { background: #f0fdf4; border-left: 4px solid #22c55e; color: #15803d; }
    .result-msg.rejected { background: #fef2f2; border-left: 4px solid #ef4444; color: #b91c1c; }
    .days-left { margin-top: 6px; font-size: 13px; }
    .history { margin-top: 20px; border-top: 2px dashed #e2e8f0; padding-top: 15px; }
    h3 { color: #475569; font-size: 15px; margin-bottom: 10px; }
    .history-row { display: flex; justify-content: space-between; align-items: center; padding: 8px; border-radius: 6px; margin-bottom: 6px; font-size: 13px; }
    .history-row.accepted { background: #f0fdf4; }
    .history-row.rejected { background: #fef2f2; }
    .badge { padding: 3px 8px; border-radius: 12px; font-size: 12px; font-weight: bold; background: rgba(0,0,0,0.08); }
  `]
})
export class ReturnComponent {
  barcode = '';
  articleNumber = '';
  cellLocation = '';
  issuedDate = '';
  isLoading = false;
  resultMessage = '';
  resultStatus = '';
  daysLeft: number | null = null;
  history: any[] = [];

  constructor(private http: HttpClient, private cdr: ChangeDetectorRef) {}

  isFormValid() {
    return !!(this.barcode.trim() && this.cellLocation.trim() && this.issuedDate);
  }

  processReturn() {
    this.isLoading = true;
    this.resultMessage = '';
    const body = {
      barcode: this.barcode,
      articleNumber: this.articleNumber,
      cellLocation: this.cellLocation,
      issuedDate: new Date(this.issuedDate).toISOString()
    };

    this.http.post<any>('/api/return', body).subscribe({
      next: (res) => {
        this.resultMessage = res.message;
        this.resultStatus = res.status;
        this.daysLeft = res.daysLeft ?? null;
        this.isLoading = false;
        this.cdr.detectChanges();
      },
      error: (err) => {
        this.resultMessage = err.error?.message || 'Ошибка при оформлении возврата';
        this.resultStatus = err.error?.status || 'rejected';
        this.daysLeft = null;
        this.isLoading = false;
        this.cdr.detectChanges();
      }
    });
  }

  loadHistory() {
    this.http.get<any[]>(`/api/return/${encodeURIComponent(this.barcode)}`).subscribe({
      next: (data) => { this.history = data; this.cdr.detectChanges(); },
      error: () => { this.history = []; this.cdr.detectChanges(); }
    });
  }
}
