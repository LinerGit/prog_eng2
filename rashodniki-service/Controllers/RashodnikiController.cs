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

        public RashodnikiController(AppDbContext context)
        {
            _context = context;
        }

        // Посмотреть все расходники
        [HttpGet]
        public async Task<IActionResult> GetAll()
        {
            var list = await _context.Rashodniki.ToListAsync();
            return Ok(list);
        }

        // Добавить расходники на склад
        [HttpPost("add")]
        public async Task<IActionResult> Add([FromBody] Rashodnik rashodnik)
        {
            // Если такой тип уже есть — просто увеличиваем количество
            var existing = await _context.Rashodniki
                .FirstOrDefaultAsync(r => r.PackageType == rashodnik.PackageType);

            if (existing != null)
            {
                existing.Count += rashodnik.Count;
            }
            else
            {
                _context.Rashodniki.Add(rashodnik);
            }

            await _context.SaveChangesAsync();
            return Ok(new { message = "Расходники добавлены" });
        }

        // Потратить 1 расходник (вызывается при выдаче товара)
        [HttpPost("use")]
        public async Task<IActionResult> Use([FromBody] UseRequest request)
        {
            var rashodnik = await _context.Rashodniki
                .FirstOrDefaultAsync(r => r.PackageType == request.PackageType);

            if (rashodnik == null)
                return NotFound(new { message = "Такой тип расходника не найден" });

            if (rashodnik.Count <= 0)
                return BadRequest(new { message = "Расходники закончились!" });

            rashodnik.Count -= 1;
            await _context.SaveChangesAsync();

            return Ok(new { message = "Расходник использован", remaining = rashodnik.Count });
        }
    }

    // Простой класс для запроса на использование
    public class UseRequest
    {
        public string PackageType { get; set; } = string.Empty;
    }
}
