using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using RashodnikiService.Data;
using RashodnikiService.Models;

namespace RashodnikiService.Controllers
{
    [ApiController]
    [Route("api/rashodniki")]
    public class RashodnikiController : ControllerBase
    {
        private readonly AppDbContext _context;
        private readonly IHttpClientFactory _http;
        private readonly IConfiguration _config;

        public RashodnikiController(AppDbContext context, IHttpClientFactory http, IConfiguration config)
        {
            _context = context; _http = http; _config = config;
        }

        // Просмотр остатков — публичный
        [HttpGet]
        public async Task<IActionResult> GetAll()
        {
            var list = await _context.Rashodniki.ToListAsync();
            return Ok(list);
        }

        // Добавить расходники — JWT
        [HttpPost("add")]
        public async Task<IActionResult> Add([FromBody] Rashodnik rashodnik)
        {
            var (ok, errorResult) = await ValidateUserJwt();
            if (!ok) return errorResult!;

            var existing = await _context.Rashodniki
                .FirstOrDefaultAsync(r => r.PackageType == rashodnik.PackageType);
            if (existing != null)
                existing.Count += rashodnik.Count;
            else
                _context.Rashodniki.Add(rashodnik);

            await _context.SaveChangesAsync();
            return Ok(new { message = "Расходники добавлены" });
        }

        // Списать расходники — JWT; count опционален (по умолчанию 1)
        [HttpPost("use")]
        public async Task<IActionResult> Use([FromBody] UseRequest request)
        {
            var (ok, errorResult) = await ValidateUserJwt();
            if (!ok) return errorResult!;

            var amount = request.Count > 0 ? request.Count : 1;

            var rashodnik = await _context.Rashodniki
                .FirstOrDefaultAsync(r => r.PackageType == request.PackageType);
            if (rashodnik == null)
                return NotFound(new { message = "Такой тип расходника не найден" });
            if (rashodnik.Count < amount)
                return BadRequest(new { message = $"Недостаточно расходников. Есть: {rashodnik.Count}, нужно: {amount}" });

            rashodnik.Count -= amount;
            await _context.SaveChangesAsync();
            return Ok(new { message = "Расходники списаны", remaining = rashodnik.Count });
        }

        // ─── Helper ──────────────────────────────────────────────────────────

        private async Task<(bool ok, IActionResult? error)> ValidateUserJwt()
        {
            var authHeader = Request.Headers["Authorization"].ToString();
            if (!authHeader.StartsWith("Bearer "))
                return (false, Unauthorized(new { message = "Требуется авторизация (Bearer JWT)" }));

            var client = _http.CreateClient();
            var usersUrl = _config["Services:UsersService"] ?? "http://users-service:8080";
            var req = new HttpRequestMessage(HttpMethod.Get, $"{usersUrl}/api/auth/validate-token");
            req.Headers.Add("Authorization", authHeader);
            var resp = await client.SendAsync(req);

            if (!resp.IsSuccessStatusCode)
                return (false, Unauthorized(new { message = "Недействительный токен" }));

            return (true, null);
        }
    }

    public class UseRequest
    {
        public string PackageType { get; set; } = string.Empty;
        public int Count { get; set; } = 1;
    }
}
