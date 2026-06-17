import { Component, EventEmitter, Output, ChangeDetectorRef } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [FormsModule, CommonModule],
  template: `
    <div class="login-card">
      <h2>{{ isRegister ? '📝 Регистрация' : '🔑 Вход' }}</h2>
      
      <div class="form-group">
        <label>Логин</label>
        <input [(ngModel)]="username" placeholder="Введите логин" class="form-control">
      </div>

      <div class="form-group">
        <label>Пароль</label>
        <input [(ngModel)]="password" type="password" placeholder="Введите пароль" class="form-control" (input)="checkPassword()">
        <div *ngIf="isRegister" class="hint" [class.valid]="isPasswordValid">
          Минимум 6 символов, буквы и цифры
        </div>
      </div>

      <button (click)="onSubmit()" class="btn-submit" [disabled]="isRegister && !isPasswordValid">
        {{ isRegister ? 'Зарегистрироваться' : 'Войти' }}
      </button>

      <p class="switch-mode">
        {{ isRegister ? 'Уже есть аккаунт?' : 'Нет аккаунта?' }}
        <a href="#" (click)="toggleMode($event)">
          {{ isRegister ? 'Войти' : 'Создать' }}
        </a>
      </p>

      <p *ngIf="message" class="message" [class.error]="isError">{{ message }}</p>
    </div>
  `,
  styles: [`
    .login-card { max-width: 350px; margin: 80px auto; padding: 30px; background: white; border-radius: 12px; box-shadow: 0 10px 25px rgba(0,0,0,0.2); font-family: sans-serif; }
    h2 { text-align: center; color: #2c3e50; margin-bottom: 20px; }
    
    .form-group { margin-bottom: 15px; }
    label { display: block; margin-bottom: 5px; color: #7f8c8d; font-size: 14px; }
    .form-control { width: 100%; padding: 10px; border: 2px solid #ecf0f1; border-radius: 6px; box-sizing: border-box; }
    .form-control:focus { border-color: #3498db; outline: none; }

    .btn-submit { width: 100%; padding: 12px; background: #2ecc71; color: white; border: none; border-radius: 6px; font-weight: bold; cursor: pointer; margin-top: 10px; }
    .btn-submit:disabled { background: #95a5a6; cursor: not-allowed; }
    .btn-submit:hover:not(:disabled) { background: #27ae60; }

    .switch-mode { text-align: center; margin-top: 15px; font-size: 14px; }
    .switch-mode a { color: #3498db; text-decoration: none; font-weight: bold; }
    
    .message { text-align: center; margin-top: 15px; padding: 10px; border-radius: 4px; background: #dff9fb; color: #130f40; }
    .message.error { background: #ff7979; color: white; }

    .hint { font-size: 12px; color: #e74c3c; margin-top: 5px; }
    .hint.valid { color: #2ecc71; }
  `]
})
export class LoginComponent {
  isRegister = false;
  username = '';
  password = '';
  
  message = '';
  isError = false;
  isPasswordValid = false;

  @Output() loginSuccess = new EventEmitter<void>();

  constructor(
    private authService: AuthService,
    private cdr: ChangeDetectorRef 
  ) {}

  toggleMode(event: Event) {
    event.preventDefault();
    this.isRegister = !this.isRegister;
    this.message = '';
    this.password = '';
    this.checkPassword();
  }

  checkPassword() {
    const regex = /^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d]{6,}$/;
    this.isPasswordValid = regex.test(this.password);
  }

  onSubmit() {
    this.message = 'Загрузка...';
    this.isError = false;

    if (this.isRegister) {
      this.authService.register(this.username, this.password).subscribe({
        next: (res) => {
          this.message = res.message; 
          this.isRegister = false; 
          this.password = '';
          
          this.cdr.detectChanges();
        },
        error: (err) => {
          this.isError = true;
          this.message = err.error?.message || 'Ошибка регистрации';
          
          this.cdr.detectChanges();
        }
      });
    } else {
      this.authService.login(this.username, this.password).subscribe({
        next: () => {
          this.loginSuccess.emit();
        },
        error: () => {
          this.isError = true;
          this.message = 'Неверный логин или пароль';
          
          this.cdr.detectChanges();
        }
      });
    }
  }
}