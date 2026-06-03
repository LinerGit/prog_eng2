using Microsoft.AspNetCore.Mvc;
using AddTovarService.Data;
using AddTovarService.Models;
using System.Text;

namespace AddTovarService.Controllers
{
    [Route("api/[controller]")]
    [ApiController]
    public class ProductController : ControllerBase
    {
        private readonly AddTovarDbContext _context;
        private readonly IHttpClientFactory _httpClientFactory;
        private readonly IConfiguration _config;

        public ProductController(AddTovarDbContext context, IHttpClientFactory httpClientFactory, IConfiguration config)
        {
            _context = context; _httpClientFactory = httpClientFactory; _config = config;
        }

        [HttpPost("add")]
        public async Task<IActionResult> Add([FromBody] Product product)
        {
            if (!Request.Headers.ContainsKey("Authorization"))
                return Unauthorized(new { message = "Нужна авторизация!" });

            var authHeader = Request.Headers["Authorization"].ToString();
            var usersServiceUrl = _config["Services:UsersService"];
            var client = _httpClientFactory.CreateClient();

            if (authHeader.StartsWith("Bearer "))
            {
                // Лаба 4: JWT
                var req = new HttpRequestMessage(HttpMethod.Get,
                    $"{usersServiceUrl}/api/auth/validate-token?requiredRole=admin");
                req.Headers.Add("Authorization", authHeader);
                var resp = await client.SendAsync(req);
                if (!resp.IsSuccessStatusCode)
                    return StatusCode((int)resp.StatusCode, new { message = "Доступ запрещён" });

                var username = System.Security.Claims.ClaimsPrincipal.Current?.Identity?.Name ?? "unknown";
                return await SaveProduct(product, username);
            }
            else if (authHeader.StartsWith("Basic "))
            {
                // Лаба 2: Basic Auth (обратная совместимость)
                var encoded = authHeader.Substring("Basic ".Length).Trim();
                var decoded = Encoding.GetEncoding("iso-8859-1").GetString(Convert.FromBase64String(encoded));
                var parts = decoded.Split(':');
                if (parts.Length != 2) return Unauthorized(new { message = "Ошибка авторизации" });

                var resp = await client.GetAsync(
                    $"{usersServiceUrl}/api/auth/validate?username={Uri.EscapeDataString(parts[0])}&password={Uri.EscapeDataString(parts[1])}&requiredRole=admin");

                if (resp.StatusCode == System.Net.HttpStatusCode.Unauthorized)
                    return Unauthorized(new { message = "Неверный логин или пароль" });
                if (!resp.IsSuccessStatusCode)
                    return StatusCode(403, new { message = "Доступ запрещен. Только Администратор!" });

                return await SaveProduct(product, parts[0]);
            }

            return Unauthorized(new { message = "Неверный формат авторизации" });
        }

        private async Task<IActionResult> SaveProduct(Product product, string username)
        {
            product.ReceivedDate = DateTime.UtcNow;
            product.ReceivedByEmployee = username;
            _context.Products.Add(product);
            await _context.SaveChangesAsync();
            return Ok(new { message = "Товар успешно добавлен!", id = product.Id });
        }

        [HttpGet("search")]
        public IActionResult Search([FromQuery] string searchType, [FromQuery] string searchValue)
        {
            if (string.IsNullOrWhiteSpace(searchType) || searchValue == null)
                return BadRequest(new { message = "Укажите тип и значение поиска" });
            var val = searchValue.ToLower().Trim();
            var query = _context.Products.AsQueryable();
            var product = searchType switch
            {
                "barcode"        => query.FirstOrDefault(p => p.Barcode != null && p.Barcode.ToLower() == val),
                "storageCell"    => query.FirstOrDefault(p => p.CellLocation.ToLower() == val),
                "productBarcode" => query.FirstOrDefault(p => p.ProductBarcode != null && p.ProductBarcode.ToLower() == val),
                "article"        => query.FirstOrDefault(p => p.ArticleNumber != null && p.ArticleNumber.ToLower() == val),
                _                => null
            };
            return Ok(product);
        }

        [HttpGet("{id}")]
        public IActionResult GetById(int id)
        {
            var product = _context.Products.FirstOrDefault(p => p.Id == id);
            if (product == null) return NotFound();
            return Ok(product);
        }

        [HttpPut("{id}/issue")]
        public async Task<IActionResult> Issue(int id, [FromBody] IssueRequest req)
        {
            var product = _context.Products.FirstOrDefault(p => p.Id == id);
            if (product == null) return NotFound(new { message = "Товар не найден" });
            if (product.IssuedDate != null) return BadRequest(new { message = "Товар уже выдан" });
            product.IssuedDate = DateTime.UtcNow;
            product.IssuedByEmployee = req.EmployeeName;
            product.WeightIssued = req.WeightIssued;
            await _context.SaveChangesAsync();
            return Ok(new { message = "Выдача зафиксирована", issuedDate = product.IssuedDate });
        }
    }

    public class IssueRequest
    {
        public string EmployeeName { get; set; } = string.Empty;
        public double? WeightIssued { get; set; }
    }
}
