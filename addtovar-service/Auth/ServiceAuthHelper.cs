using System.Text;

namespace AddTovarService.Auth
{
    /// <summary>
    /// Проверяет Basic Auth для межсервисных вызовов.
    /// Credentials берутся из конфига: ServiceAuth:Username / ServiceAuth:Password
    /// </summary>
    public static class ServiceAuthHelper
    {
        public static bool IsValidServiceRequest(HttpRequest request, IConfiguration config)
        {
            var authHeader = request.Headers["Authorization"].ToString();
            if (!authHeader.StartsWith("Basic ")) return false;

            var encoded = authHeader["Basic ".Length..].Trim();
            string decoded;
            try { decoded = Encoding.UTF8.GetString(Convert.FromBase64String(encoded)); }
            catch { return false; }

            var parts = decoded.Split(':', 2);
            if (parts.Length != 2) return false;

            var expectedUser = config["ServiceAuth:Username"];
            var expectedPass = config["ServiceAuth:Password"];
            return parts[0] == expectedUser && parts[1] == expectedPass;
        }
    }
}
