using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Infrastructure;
using AddTovarService.Data;

#nullable disable

namespace AddTovarService.Migrations
{
    [DbContext(typeof(AddTovarDbContext))]
    partial class AddTovarDbContextModelSnapshot : ModelSnapshot
    {
        protected override void BuildModel(ModelBuilder modelBuilder)
        {
            modelBuilder.HasAnnotation("ProductVersion", "8.0.0")
                .HasAnnotation("Relational:MaxIdentifierLength", 63);

            modelBuilder.Entity("AddTovarService.Models.Product", b =>
            {
                b.Property<int>("Id").ValueGeneratedOnAdd().HasColumnType("integer");
                b.Property<string>("ProductName").IsRequired().HasColumnType("text");
                b.Property<string>("CellLocation").IsRequired().HasColumnType("text");
                b.Property<string>("Barcode").HasColumnType("text");
                b.Property<string>("ProductBarcode").HasColumnType("text");
                b.Property<string>("ArticleNumber").HasColumnType("text");
                b.Property<double>("WeightReceived").HasColumnType("double precision");
                b.Property<DateTime>("ReceivedDate").HasColumnType("timestamp with time zone");
                b.Property<string>("ReceivedByEmployee").IsRequired().HasColumnType("text");
                b.Property<DateTime?>("IssuedDate").HasColumnType("timestamp with time zone");
                b.Property<string>("IssuedByEmployee").HasColumnType("text");
                b.Property<double?>("WeightIssued").HasColumnType("double precision");
                b.HasKey("Id");
                b.ToTable("Products");
            });
        }
    }
}
