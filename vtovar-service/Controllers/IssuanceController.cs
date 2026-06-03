using Microsoft.AspNetCore.Mvc;
using System.Net.Http.Json;
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

        [HttpGet("search")]
        public async Task<IActionResult> Search([FromQuery] string searchType, [FromQuery] string searchValue)
        {
            if (string.IsNullOrWhiteSpace(searchType) || string.IsNullOrWhiteSpace(searchValue))
                return BadRequest(new { message = "Укажите тип и значение поиска" });

            var client = _http.CreateClient();
            var resp = await client.GetAsync(
                $"{_config["Services:AddTovarService"]}/api/product/search?searchType={Uri.EscapeDataString(searchType)}&searchValue={Uri.EscapeDataString(searchValue)}");

            if (!resp.IsSuccessStatusCode)
                return StatusCode((int)resp.StatusCode, new { message = "Ошибка при поиске товара" });

            return Ok(await resp.Content.ReadFromJsonAsync<ProductDto>());
        }

        [HttpPost("issue")]
        public async Task<IActionResult> Issue([FromBody] IssueRequest req)
        {
            if (req.ProductId <= 0 || string.IsNullOrWhiteSpace(req.EmployeeName))
                return BadRequest(new { message = "Укажите ProductId и имя сотрудника" });

            var client = _http.CreateClient();
            var addTovarUrl = _config["Services:AddTovarService"];

            var checkResp = await client.GetAsync($"{addTovarUrl}/api/product/{req.ProductId}");
            if (!checkResp.IsSuccessStatusCode)
                return NotFound(new { message = "Товар не найден" });

            var product = await checkResp.Content.ReadFromJsonAsync<ProductDto>();
            if (product?.IssuedDate != null)
                return BadRequest(new { message = "Товар уже выдан" });

            var issueResp = await client.PutAsJsonAsync(
                $"{addTovarUrl}/api/product/{req.ProductId}/issue",
                new { employeeName = req.EmployeeName, weightIssued = req.WeightIssued });

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

            // Лаба 3: публикуем событие вместо HTTP-вызова rashodniki
            _rabbit.Publish("item.issued", new
            {
                productId = req.ProductId,
                employeeName = req.EmployeeName,
                packageType = req.PackageType ?? "default",
                issuedAt = issuance.IssuedAt
            });

            return Ok(new { message = "Выдача зафиксирована", issuanceId = issuance.Id });
        }

        [HttpGet("history")]
        public IActionResult History()
        {
            var list = _context.Issuances.OrderByDescending(i => i.IssuedAt).ToList();
            return Ok(list);
        }
    }
}
