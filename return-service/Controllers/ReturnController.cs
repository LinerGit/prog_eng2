using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using ReturnService.Data;
using ReturnService.Models;

namespace ReturnService.Controllers
{
    [ApiController]
    [Route("api/return")]
    public class ReturnController : ControllerBase
    {
        private readonly AppDbContext _context;
        private readonly IHttpClientFactory _http;
        private readonly IConfiguration _config;
        private const int ReturnDeadlineDays = 14;

        public ReturnController(AppDbContext context, IHttpClientFactory http, IConfiguration config)
        {
            _context = context; _http = http; _config = config;
        }

        // Оформить возврат — требует JWT
        [HttpPost]
        public async Task<IActionResult> ProcessReturn([FromBody] ReturnRequest request)
        {
            var (ok, errorResult) = await ValidateUserJwt();
            if (!ok) return errorResult!;

            var today = DateTime.UtcNow;
            var daysSinceIssued = (today - request.IssuedDate).TotalDays;

            if (daysSinceIssued > ReturnDeadlineDays)
            {
                var rejected = new ReturnRecord
                {
                    Barcode = request.Barcode,
                    ArticleNumber = request.ArticleNumber,
                    CellLocation = request.CellLocation,
                    IssuedDate = request.IssuedDate,
                    ReturnDate = today,
                    Status = "rejected"
                };
                _context.Returns.Add(rejected);
                await _context.SaveChangesAsync();
                return BadRequest(new
                {
                    message = $"Срок возврата истёк. Прошло {(int)daysSinceIssued} дней, максимум {ReturnDeadlineDays}.",
                    status = "rejected"
                });
            }

            var accepted = new ReturnRecord
            {
                Barcode = request.Barcode,
                ArticleNumber = request.ArticleNumber,
                CellLocation = request.CellLocation,
                IssuedDate = request.IssuedDate,
                ReturnDate = today,
                Status = "accepted"
            };
            _context.Returns.Add(accepted);
            await _context.SaveChangesAsync();
            return Ok(new
            {
                message = "Возврат успешно оформлен.",
                status = "accepted",
                daysLeft = ReturnDeadlineDays - (int)daysSinceIssued
            });
        }

        // Все возвраты — требует JWT
        [HttpGet]
        public async Task<IActionResult> GetAll()
        {
            var (ok, errorResult) = await ValidateUserJwt();
            if (!ok) return errorResult!;

            return Ok(await _context.Returns.ToListAsync());
        }

        // Возвраты по штрихкоду — требует JWT
        [HttpGet("{barcode}")]
        public async Task<IActionResult> GetByBarcode(string barcode)
        {
            var (ok, errorResult) = await ValidateUserJwt();
            if (!ok) return errorResult!;

            var returns = await _context.Returns.Where(r => r.Barcode == barcode).ToListAsync();
            if (!returns.Any())
                return NotFound(new { message = "Возвраты по этому штрихкоду не найдены" });
            return Ok(returns);
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
}
