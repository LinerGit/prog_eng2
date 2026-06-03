using Microsoft.EntityFrameworkCore;
using AddTovarService.Models;

namespace AddTovarService.Data
{
    public class AddTovarDbContext : DbContext
    {
        public AddTovarDbContext(DbContextOptions<AddTovarDbContext> options) : base(options) { }

        public DbSet<Product> Products { get; set; }
    }
}
