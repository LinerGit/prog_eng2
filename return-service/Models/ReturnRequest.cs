namespace ReturnService.Models
{
    // Это то, что присылает фронт когда хочет оформить возврат
    public class ReturnRequest
    {
        public string Barcode { get; set; } = string.Empty;
        public string ArticleNumber { get; set; } = string.Empty;
        public string CellLocation { get; set; } = string.Empty;
        public DateTime IssuedDate { get; set; } // дата когда товар выдали
    }
}
