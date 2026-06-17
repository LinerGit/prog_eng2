using Microsoft.AspNetCore.Mvc;
using System.Net.Http.Json;
using System.Text;
using VTovarService.Data;
using VTovarService.Messaging;
using VTovarService.Models;

namespace VTovarService.Controllers
{
    [Route("api/[controller]")]
    [ApiController]
    public class IssuanceController : ControllerBase
    {
        private readonly VTovarDbContext _context;
        private readonly IHttpClientFactory _http;
        private readonly IConfiguration _config;
        private readonly RabbitPublisher _rabbit;

        public IssuanceController(VTovarDbContext context, IHttpClientFactory http,
            IConfiguration config, RabbitPublisher rabbit)
        {
            _context = context; _http = http; _config = config; _rabbit = rabbit;
        }

        // Поиск товара — требует JWT
        [HttpGet("search")]
        public async Task<IActionResult> Search([FromQuery] string searchType, [FromQuery] string searchValue)
        {
            var (ok, errorResult) = await ValidateUserJwt();
            if (!ok) return errorResult!;

            if (string.IsNullOrWhiteSpace(searchType) || string.IsNullOrWhiteSpace(searchValue))
                return BadRequest(new { message = "Укажите тип и значение поиска" });

            var client = _http.CreateClient();
            var req = new HttpRequestMessage(HttpMethod.Get,
                $"{_config["Services:AddTovarService"]}/api/product/search?searchType={Uri.EscapeDataString(searchType)}&searchValue={Uri.EscapeDataString(searchValue)}");
            req.Headers.Add("Authorization", ServiceBasicHeader());
            var resp = await client.SendAsync(req);

            if (!resp.IsSuccessStatusCode)
                return StatusCode((int)resp.StatusCode, new { message = "Ошибка при поиске товара" });

            return Ok(await resp.Content.ReadFromJsonAsync<ProductDto>());
        }

        // Выдача — требует JWT
        [HttpPost("issue")]
        public async Task<IActionResult> Issue([FromBody] IssueRequest req)
        {
            var (ok, errorResult) = await ValidateUserJwt();
            if (!ok) return errorResult!;

            if (req.ProductId <= 0 || string.IsNullOrWhiteSpace(req.EmployeeName))
                return BadRequest(new { message = "Укажите ProductId и имя сотрудника" });

            var client = _http.CreateClient();
            var addTovarUrl = _config["Services:AddTovarService"];

            var checkReq = new HttpRequestMessage(HttpMethod.Get, $"{addTovarUrl}/api/product/{req.ProductId}");
            checkReq.Headers.Add("Authorization", ServiceBasicHeader());
            var checkResp = await client.SendAsync(checkReq);
            if (!checkResp.IsSuccessStatusCode)
                return NotFound(new { message = "Товар не найден" });

            var product = await checkResp.Content.ReadFromJsonAsync<ProductDto>();
            if (product?.IssuedDate != null)
                return BadRequest(new { message = "Товар уже выдан" });

            var issueReq = new HttpRequestMessage(HttpMethod.Put, $"{addTovarUrl}/api/product/{req.ProductId}/issue");
            issueReq.Headers.Add("Authorization", ServiceBasicHeader());
            issueReq.Content = JsonContent.Create(new { employeeName = req.EmployeeName, weightIssued = req.WeightIssued });
            var issueResp = await client.SendAsync(issueReq);

            if (!issueResp.IsSuccessStatusCode)
                return StatusCode(500, new { message = "Не удалось зафиксировать выдачу в addtovar" });

            var issuance = new Issuance
            {
                ProductId = req.ProductId,
                IssuedByEmployee = req.EmployeeName,
                IssuedAt = DateTime.UtcNow
            };
            _context.Issuances.Add(issuance);
            await _context.SaveChangesAsync();

            _rabbit.Publish("item.issued", new
            {
                productId = req.ProductId,
                employeeName = req.EmployeeName,
                packageType = req.PackageType ?? "default",
                issuedAt = issuance.IssuedAt
            });

            return Ok(new { message = "Выдача зафиксирована", issuanceId = issuance.Id });
        }

        // История выдач — требует JWT
        [HttpGet("history")]
        public async Task<IActionResult> History()
        {
            var (ok, errorResult) = await ValidateUserJwt();
            if (!ok) return errorResult!;

            var list = _context.Issuances.OrderByDescending(i => i.IssuedAt).ToList();
            return Ok(list);
        }

        // ─── Helpers ─────────────────────────────────────────────────────────

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

        private string ServiceBasicHeader()
        {
            var user = _config["ServiceAuth:Username"] ?? "svc";
            var pass = _config["ServiceAuth:Password"] ?? "svc-secret-2024";
            return "Basic " + Convert.ToBase64String(Encoding.UTF8.GetBytes($"{user}:{pass}"));
        }
    }
}
