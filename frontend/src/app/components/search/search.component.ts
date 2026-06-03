import { Component, ChangeDetectorRef } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { ProductService } from '../../services/product.service';
import { AuthService } from '../../services/auth.service';
import { HttpClient, HttpHeaders } from '@angular/common/http';

@Component({
  selector: 'app-search',
  standalone: true,
  imports: [FormsModule, CommonModule],
  template: `
    <div class="search-card">
      <h1>📦 Поиск товаров</h1>
      
      <div *ngIf="authService.currentUser?.role === 'admin'" class="admin-panel">
        <div class="admin-header"><span>🛡️ Добавление нового товара</span></div>
        <div class="admin-form">
          <div class="form-row">
            <input [(ngModel)]="newProduct.productName" placeholder="Наименование товара" class="form-control admin-input">
            <input [(ngModel)]="newProduct.cellLocation" placeholder="Ячейка (например A-01)" class="form-control admin-input">
          </div>
          <div class="form-row">
            <input [(ngModel)]="newProduct.weightReceived" type="number" placeholder="Вес (кг)" class="form-control admin-input">
          </div>
          <div class="form-row id-selector">
            <select [(ngModel)]="selectedIdType" class="form-control admin-select">
              <option value="barcode">Штрихкод</option>
              <option value="article">Артикул</option>
              <option value="productBarcode">Баркод товара</option>
            </select>
            <input [(ngModel)]="selectedIdValue" placeholder="Введите значение..." class="form-control admin-input">
          </div>
          <button (click)="addProduct()" class="btn-add" [disabled]="!isFormValid()">💾 Сохранить в базу</button>
          <span *ngIf="addMessage" class="status-msg" [class.error]="addMessage.includes('Ошибка')">{{ addMessage }}</span>
        </div>
      </div>

      <div class="search-form">
        <div class="form-group">
          <label>Параметр поиска:</label>
          <select [(ngModel)]="searchType" class="form-control">
            <option value="barcode">Штрихкод</option>
            <option value="storageCell">Ячейка на складе</option>
            <option value="productBarcode">Баркод товара</option>
            <option value="article">Артикул</option>
          </select>
        </div>
        <div class="form-group">
          <label>Значение:</label>
          <input type="text" [(ngModel)]="searchValue" class="form-control" placeholder="Например: 1234567890123" (keyup.enter)="onSearch()">
        </div>
        <button class="btn-search" (click)="onSearch()" [disabled]="isLoading">{{ isLoading ? '⏳ Ищем...' : '🔍 Найти' }}</button>
      </div>

      <div *ngIf="searchResult" class="result-card">
        <h3>✅ Товар найден</h3>
        <div class="info-row"><span class="label">Название:</span> <span class="value">{{ searchResult.productName }}</span></div>
        <div class="info-row"><span class="label">Ячейка:</span> <span class="value highlight">{{ searchResult.cellLocation }}</span></div>
        <div class="info-row"><span class="label">Вес:</span> <span class="value">{{ searchResult.weightReceived }} кг</span></div>
        <div class="info-row"><span class="label">Принял:</span> <span class="value">{{ searchResult.receivedByEmployee }}</span></div>
        <div class="info-row"><span class="label">Дата:</span> <span class="value">{{ searchResult.receivedDate | date:'dd.MM.yyyy' }}</span></div>
        <hr>
        <div class="info-row" *ngIf="searchResult.barcode"><span class="label">Штрихкод:</span> <span class="value">{{ searchResult.barcode }}</span></div>
        <div class="info-row" *ngIf="searchResult.articleNumber"><span class="label">Артикул:</span> <span class="value">{{ searchResult.articleNumber }}</span></div>
        <div class="info-row" *ngIf="searchResult.productBarcode"><span class="label">Баркод:</span> <span class="value">{{ searchResult.productBarcode }}</span></div>
      </div>
      <div *ngIf="notFound" class="not-found">❌ Товар не найден. Проверьте данные.</div>
    </div>
  `,
  styles: [`
    .search-card { background: white; max-width: 600px; margin: 40px auto; padding: 30px; border-radius: 12px; box-shadow: 0 10px 25px rgba(0,0,0,0.2); font-family: sans-serif; }
    h1 { text-align: center; color: #2c3e50; font-size: 24px; }
    .admin-panel { background: #f0f9ff; border: 1px solid #bae6fd; border-radius: 8px; padding: 15px; margin-bottom: 25px; }
    .admin-header { font-weight: bold; color: #0369a1; margin-bottom: 15px; border-bottom: 1px solid #bae6fd; padding-bottom: 5px; }
    .form-row { display: flex; gap: 10px; margin-bottom: 10px; }
    .admin-input { font-size: 14px; padding: 8px; }
    .admin-select { font-size: 14px; padding: 8px; width: 40%; }
    .btn-add { background: #0ea5e9; color: white; border: none; padding: 10px 20px; border-radius: 6px; cursor: pointer; font-weight: bold; width: 100%; }
    .btn-add:disabled { background: #cbd5e1; cursor: not-allowed; }
    .status-msg { display: block; text-align: center; margin-top: 10px; color: #16a34a; font-weight: bold; }
    .status-msg.error { color: #dc2626; }
    .form-group { margin-bottom: 15px; }
    label { display: block; margin-bottom: 5px; font-weight: 600; color: #64748b; }
    .form-control { width: 100%; padding: 12px; border: 2px solid #e2e8f0; border-radius: 8px; box-sizing: border-box; }
    .btn-search { width: 100%; padding: 14px; background: #6366f1; color: white; border: none; border-radius: 8px; font-weight: bold; cursor: pointer; margin-top: 10px; }
    .result-card { margin-top: 25px; padding: 20px; background: #f8fafc; border-radius: 8px; border-left: 5px solid #22c55e; }
    .info-row { display: flex; justify-content: space-between; margin-bottom: 8px; border-bottom: 1px solid #e2e8f0; padding-bottom: 5px; }
    .value.highlight { color: #6366f1; font-weight: bold; font-size: 1.1em; }
    .not-found { margin-top: 20px; padding: 15px; background: #fef2f2; color: #dc2626; border-radius: 8px; text-align: center; border: 1px solid #fecaca; }
  `]
})
export class SearchComponent {
  searchType = 'barcode';
  searchValue = '';
  searchResult: any = null;
  notFound = false;
  isLoading = false;
  addMessage = '';
  newProduct: any = { productName: '', cellLocation: '', weightReceived: null };
  selectedIdType = 'barcode';
  selectedIdValue = '';

  constructor(
    private productService: ProductService,
    private cdr: ChangeDetectorRef,
    public authService: AuthService,
    private http: HttpClient
  ) {}

  onSearch(): void {
    if (!this.searchValue.trim()) return;
    this.isLoading = true; this.notFound = false; this.searchResult = null;
    this.productService.searchProduct({ searchType: this.searchType, searchValue: this.searchValue })
      .subscribe({
        next: (data) => { this.isLoading = false; this.searchResult = data || null; this.notFound = !data; this.cdr.detectChanges(); },
        error: () => { this.isLoading = false; this.notFound = true; this.cdr.detectChanges(); }
      });
  }

  isFormValid(): boolean {
    return !!(this.newProduct.productName && this.newProduct.cellLocation && this.newProduct.weightReceived && this.selectedIdValue);
  }

  addProduct() {
    this.addMessage = '⏳ Сохранение...';
    const productToSend: any = {
      productName: this.newProduct.productName,
      cellLocation: this.newProduct.cellLocation,
      weightReceived: this.newProduct.weightReceived
    };
    if (this.selectedIdType === 'barcode') productToSend.barcode = this.selectedIdValue;
    if (this.selectedIdType === 'article') productToSend.articleNumber = this.selectedIdValue;
    if (this.selectedIdType === 'productBarcode') productToSend.productBarcode = this.selectedIdValue;

    // Лаба 4: Bearer JWT
    const headers = new HttpHeaders({ 'Authorization': this.authService.getAuthHeader() });
    this.http.post('/api/product/add', productToSend, { headers }).subscribe({
      next: () => {
        this.addMessage = `✅ Товар "${productToSend.productName}" добавлен!`;
        this.newProduct = { productName: '', cellLocation: '', weightReceived: null };
        this.selectedIdValue = '';
        this.cdr.detectChanges();
      },
      error: (err) => { this.addMessage = '❌ Ошибка при добавлении!'; this.cdr.detectChanges(); }
    });
  }
}
