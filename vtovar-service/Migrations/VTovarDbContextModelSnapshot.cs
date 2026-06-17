using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Infrastructure;
using VTovarService.Data;

#nullable disable

namespace VTovarService.Migrations
{
    [DbContext(typeof(VTovarDbContext))]
    partial class VTovarDbContextModelSnapshot : ModelSnapshot
    {
        protected override void BuildModel(ModelBuilder modelBuilder)
        {
            modelBuilder.HasAnnotation("ProductVersion", "8.0.0")
                .HasAnnotation("Relational:MaxIdentifierLength", 63);

            modelBuilder.Entity("VTovarService.Models.Issuance", b =>
            {
                b.Property<int>("Id").ValueGeneratedOnAdd().HasColumnType("integer");
                b.Property<int>("ProductId").HasColumnType("integer");
                b.Property<string>("IssuedByEmployee").IsRequired().HasColumnType("text");
                b.Property<DateTime>("IssuedAt").HasColumnType("timestamp with time zone");
                b.HasKey("Id");
                b.ToTable("Issuances");
            });
        }
    }
}
