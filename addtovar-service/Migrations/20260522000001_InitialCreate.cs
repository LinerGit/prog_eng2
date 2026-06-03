using Microsoft.EntityFrameworkCore.Migrations;
using Npgsql.EntityFrameworkCore.PostgreSQL.Metadata;

#nullable disable

namespace AddTovarService.Migrations
{
    public partial class InitialCreate : Migration
    {
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateTable(
                name: "Products",
                columns: table => new
                {
                    Id = table.Column<int>(nullable: false)
                        .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                    ProductName = table.Column<string>(nullable: false, defaultValue: ""),
                    CellLocation = table.Column<string>(nullable: false, defaultValue: ""),
                    Barcode = table.Column<string>(nullable: true),
                    ProductBarcode = table.Column<string>(nullable: true),
                    ArticleNumber = table.Column<string>(nullable: true),
                    WeightReceived = table.Column<double>(nullable: false),
                    ReceivedDate = table.Column<DateTime>(nullable: false),
                    ReceivedByEmployee = table.Column<string>(nullable: false, defaultValue: ""),
                    IssuedDate = table.Column<DateTime>(nullable: true),
                    IssuedByEmployee = table.Column<string>(nullable: true),
                    WeightIssued = table.Column<double>(nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Products", x => x.Id);
                });
        }

        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(name: "Products");
        }
    }
}
