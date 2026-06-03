using Microsoft.AspNetCore.Mvc;
using Microsoft.IdentityModel.Tokens;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;
using System.Text.RegularExpressions;
using UsersService.Data;
using UsersService.Models;

namespace UsersService.Controllers
{
    [Route("api/[controller]")]
    [ApiController]
    public class AuthController : ControllerBase
    {
        private readonly UsersDbContext _context;
        private readonly IConfiguration _config;

        public AuthController(UsersDbContext context, IConfiguration config)
        {
            _context = context; _config = config;
        }

        [HttpPost("login")]
        public IActionResult Login([FromBody] LoginRequest request)
        {
            var user = _context.Users.FirstOrDefault(
                u => u.Username == request.Username && u.PasswordHash == request.Password);
            if (user == null)
                return Unauthorized(new { message = "Неверный логин или пароль" });

            var token = GenerateJwt(user);
            return Ok(new { role = user.Role, username = user.Username, token });
        }

        [HttpPost("register")]
        public IActionResult Register([FromBody] LoginRequest request)
        {
            if (_context.Users.Any(u => u.Username == request.Username))
                return BadRequest(new { message = "Такой пользователь уже существует!" });

            var regex = new Regex(@"^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d]{6,}$");
            if (!regex.IsMatch(request.Password))
                return BadRequest(new { message = "Пароль слишком простой! Нужны буквы, цифры и минимум 6 знаков." });

            _context.Users.Add(new User
            {
                Username = request.Username,
                PasswordHash = request.Password,
                Role = "employee"
            });
            _context.SaveChanges();
            return Ok(new { message = "Регистрация успешна! Теперь войдите." });
        }

        // Лаба 2: Basic Auth валидация для межсервисных вызовов (обратная совместимость)
        [HttpGet("validate")]
        public IActionResult Validate([FromQuery] string username, [FromQuery] string password, [FromQuery] string? requiredRole)
        {
            var user = _context.Users.FirstOrDefault(
                u => u.Username == username && u.PasswordHash == password);
            if (user == null)
                return Unauthorized(new { valid = false });
            if (!string.IsNullOrEmpty(requiredRole) && user.Role != requiredRole)
                return StatusCode(403, new { valid = false, message = "Недостаточно прав" });
            return Ok(new { valid = true, role = user.Role, username = user.Username });
        }

        // Лаба 4: валидация JWT токена для межсервисных вызовов
        [HttpGet("validate-token")]
        public IActionResult ValidateToken([FromQuery] string? requiredRole)
        {
            var authHeader = Request.Headers["Authorization"].ToString();
            if (!authHeader.StartsWith("Bearer "))
                return Unauthorized(new { valid = false });

            var token = authHeader.Substring("Bearer ".Length).Trim();
            try
            {
                var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(
                    _config["Jwt:Secret"] ?? "warehouse-secret-key-min-32-chars!!"));
                var handler = new JwtSecurityTokenHandler();
                var principal = handler.ValidateToken(token, new TokenValidationParameters
                {
                    ValidateIssuerSigningKey = true,
                    IssuerSigningKey = key,
                    ValidateIssuer = false,
                    ValidateAudience = false,
                    ClockSkew = TimeSpan.Zero
                }, out _);

                var role = principal.FindFirst(ClaimTypes.Role)?.Value;
                if (!string.IsNullOrEmpty(requiredRole) && role != requiredRole)
                    return StatusCode(403, new { valid = false });

                return Ok(new { valid = true, role, username = principal.Identity?.Name });
            }
            catch
            {
                return Unauthorized(new { valid = false });
            }
        }

        private string GenerateJwt(User user)
        {
            var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(
                _config["Jwt:Secret"] ?? "warehouse-secret-key-min-32-chars!!"));
            var creds = new SigningCredentials(key, SecurityAlgorithms.HmacSha256);
            var claims = new[]
            {
                new Claim(ClaimTypes.Name, user.Username),
                new Claim(ClaimTypes.Role, user.Role),
                new Claim("username", user.Username)
            };
            var jwt = new JwtSecurityToken(
                expires: DateTime.UtcNow.AddHours(8),
                claims: claims,
                signingCredentials: creds);
            return new JwtSecurityTokenHandler().WriteToken(jwt);
        }
    }
}
