import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class ProductService {
  private apiUrl = '/api/product/search';
  constructor(private http: HttpClient) {}

  searchProduct(params: any): Observable<any> {
    const httpParams = new HttpParams()
      .set('searchType', params.searchType)
      .set('searchValue', params.searchValue);

    return this.http.get<any>(this.apiUrl, { params: httpParams });
  }
}