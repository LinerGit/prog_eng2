namespace AddTovarService.Models
{
    public class Product
    {
        public int Id { get; set; }
        public string ProductName { get; set; } = string.Empty;
        public string CellLocation { get; set; } = string.Empty;
        public string? Barcode { get; set; }
        public string? ProductBarcode { get; set; }
        public string? ArticleNumber { get; set; }
        public double WeightReceived { get; set; }
        public DateTime ReceivedDate { get; set; }
        public string ReceivedByEmployee { get; set; } = string.Empty;
        public DateTime? IssuedDate { get; set; }
        public string? IssuedByEmployee { get; set; }
        public double? WeightIssued { get; set; }
    }
}
