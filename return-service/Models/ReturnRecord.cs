namespace ReturnService.Models
{
    // Запись о возврате в базе данных
    public class ReturnRecord
    {
        public int Id { get; set; }
        public string Barcode { get; set; } = string.Empty;      // штрихкод товара
        public string ArticleNumber { get; set; } = string.Empty; // артикул
        public string CellLocation { get; set; } = string.Empty;  // ячейка
        public DateTime IssuedDate { get; set; }                  // когда выдали
        public DateTime ReturnDate { get; set; }                  // когда вернули
        public string Status { get; set; } = string.Empty;        // "accepted" или "rejected"
    }
}
