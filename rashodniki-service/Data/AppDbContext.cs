using Microsoft.EntityFrameworkCore;
using RashodnikiService.Models;

namespace RashodnikiService.Data
{
    public class AppDbContext : DbContext
    {
        public AppDbContext(DbContextOptions<AppDbContext> options) : base(options) { }

        public DbSet<Rashodnik> Rashodniki { get; set; }
    }
}
