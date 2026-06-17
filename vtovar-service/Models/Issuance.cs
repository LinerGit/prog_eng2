namespace VTovarService.Models
{
    public class Issuance
    {
        public int Id { get; set; }
        public int ProductId { get; set; }
        public string IssuedByEmployee { get; set; } = string.Empty;
        public DateTime IssuedAt { get; set; }
    }

    public class ProductDto
    {
        public int Id { get; set; }
        public string ProductName { get; set; } = string.Empty;
        public string CellLocation { get; set; } = string.Empty;
        public string? Barcode { get; set; }
        public string? ArticleNumber { get; set; }
        public DateTime? IssuedDate { get; set; }
    }

    public class IssueRequest
    {
        public int ProductId { get; set; }
        public string EmployeeName { get; set; } = string.Empty;
        public double? WeightIssued { get; set; }
        public string? PackageType { get; set; }  // тип расходника (опционально)
    }
}
