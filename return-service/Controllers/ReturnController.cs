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
        private const int ReturnDeadlineDays = 14; // срок возврата — 2 недели

        public ReturnController(AppDbContext context)
        {
            _context = context;
        }

        // Оформить возврат
        [HttpPost]
        public async Task<IActionResult> ProcessReturn([FromBody] ReturnRequest request)
        {
            var today = DateTime.UtcNow;
            var daysSinceIssued = (today - request.IssuedDate).TotalDays;

            // Проверяем: прошло ли больше 14 дней с момента выдачи
            if (daysSinceIssued > ReturnDeadlineDays)
            {
                // Срок вышел — отказываем, но всё равно записываем в базу
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

            // Всё ок — принимаем возврат
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

        // Посмотреть все возвраты
        [HttpGet]
        public async Task<IActionResult> GetAll()
        {
            var returns = await _context.Returns.ToListAsync();
            return Ok(returns);
        }

        // Посмотреть возвраты по штрихкоду
        [HttpGet("{barcode}")]
        public async Task<IActionResult> GetByBarcode(string barcode)
        {
            var returns = await _context.Returns
                .Where(r => r.Barcode == barcode)
                .ToListAsync();

            if (!returns.Any())
                return NotFound(new { message = "Возвраты по этому штрихкоду не найдены" });

            return Ok(returns);
        }
    }
}
