namespace RashodnikiService.Models
{
    public class Rashodnik
    {
        public int Id { get; set; }
        public string PackageType { get; set; } = string.Empty; // тип расходника
        public int Count { get; set; }                          // сколько осталось
    }
}
