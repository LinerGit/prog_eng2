using Microsoft.EntityFrameworkCore;
using VTovarService.Models;

namespace VTovarService.Data
{
    public class VTovarDbContext : DbContext
    {
        public VTovarDbContext(DbContextOptions<VTovarDbContext> options) : base(options) { }

        public DbSet<Issuance> Issuances { get; set; }
    }
}
