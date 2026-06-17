using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Infrastructure;
using UsersService.Data;

#nullable disable

namespace UsersService.Migrations
{
    [DbContext(typeof(UsersDbContext))]
    partial class UsersDbContextModelSnapshot : ModelSnapshot
    {
        protected override void BuildModel(ModelBuilder modelBuilder)
        {
            modelBuilder.HasAnnotation("ProductVersion", "8.0.0")
                .HasAnnotation("Relational:MaxIdentifierLength", 63);

            modelBuilder.Entity("UsersService.Models.User", b =>
            {
                b.Property<int>("Id").ValueGeneratedOnAdd().HasColumnType("integer");
                b.Property<string>("Username").IsRequired().HasColumnType("text");
                b.Property<string>("PasswordHash").IsRequired().HasColumnType("text");
                b.Property<string>("Role").IsRequired().HasColumnType("text");
                b.HasKey("Id");
                b.ToTable("Users");
            });
        }
    }
}
