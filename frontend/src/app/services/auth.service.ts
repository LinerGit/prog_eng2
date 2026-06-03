import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private apiUrl = '/api/auth';
  currentUser: { username: string, role: string, token: string, basicToken: string } | null = null;

  constructor(private http: HttpClient) {}

  login(username: string, password: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/login`, { username, password }).pipe(
      tap(response => {
        this.currentUser = {
          username,
          role: response.role,
          token: response.token,           // JWT (Лаба 4)
          basicToken: btoa(username + ':' + password)  // Basic (обратная совместимость)
        };
      })
    );
  }

  register(username: string, password: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/register`, { username, password });
  }

  logout() { this.currentUser = null; }

  getAuthHeader(): string {
    return this.currentUser ? `Bearer ${this.currentUser.token}` : '';
  }
}
