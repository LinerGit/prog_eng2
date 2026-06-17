using RabbitMQ.Client;
using RabbitMQ.Client.Events;
using System.Text;
using System.Text.Json;
using Microsoft.EntityFrameworkCore;
using RashodnikiService.Data;

namespace RashodnikiService.Messaging
{
    public class IssuedConsumer : BackgroundService
    {
        private readonly IConfiguration _config;
        private readonly IServiceScopeFactory _scopeFactory;
        private IConnection? _conn;
        private IModel? _channel;

        public IssuedConsumer(IConfiguration config, IServiceScopeFactory scopeFactory)
        {
            _config = config; _scopeFactory = scopeFactory;
        }

        protected override async Task ExecuteAsync(CancellationToken stoppingToken)
        {
            // Ждём пока RabbitMQ поднимется
            await Task.Delay(10000, stoppingToken);

            var factory = new ConnectionFactory
            {
                HostName = _config["RabbitMQ:Host"] ?? "rabbitmq",
                UserName = _config["RabbitMQ:User"] ?? "guest",
                Password = _config["RabbitMQ:Pass"] ?? "guest"
            };

            _conn = factory.CreateConnection();
            _channel = _conn.CreateModel();
            _channel.ExchangeDeclare("warehouse", ExchangeType.Topic, durable: true);
            var queue = _channel.QueueDeclare("rashodniki.issued", durable: true, exclusive: false).QueueName;
            _channel.QueueBind(queue, "warehouse", "item.issued");

            var consumer = new EventingBasicConsumer(_channel);
            consumer.Received += async (_, ea) =>
            {
                try
                {
                    var json = JsonSerializer.Deserialize<JsonElement>(Encoding.UTF8.GetString(ea.Body.ToArray()));
                    var packageType = json.TryGetProperty("packageType", out var pt) ? pt.GetString() ?? "default" : "default";

                    using var scope = _scopeFactory.CreateScope();
                    var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
                    var r = await db.Rashodniki.FirstOrDefaultAsync(x => x.PackageType == packageType);
                    if (r != null && r.Count > 0) { r.Count--; await db.SaveChangesAsync(); }

                    _channel.BasicAck(ea.DeliveryTag, false);
                }
                catch { _channel.BasicNack(ea.DeliveryTag, false, true); }
            };

            _channel.BasicConsume(queue, false, consumer);
            await Task.Delay(Timeout.Infinite, stoppingToken);
        }

        public override void Dispose() { _channel?.Dispose(); _conn?.Dispose(); base.Dispose(); }
    }
}
