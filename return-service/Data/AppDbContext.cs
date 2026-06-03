using Microsoft.EntityFrameworkCore;
using ReturnService.Models;

namespace ReturnService.Data
{
    public class AppDbContext : DbContext
    {
        public AppDbContext(DbContextOptions<AppDbContext> options) : base(options) { }

        public DbSet<ReturnRecord> Returns { get; set; }
    }
}
