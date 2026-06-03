using RabbitMQ.Client;
using System.Text;
using System.Text.Json;

namespace VTovarService.Messaging
{
    public class RabbitPublisher : IDisposable
    {
        private readonly IConnection _conn;
        private readonly IModel _channel;
        private const string Exchange = "warehouse";

        public RabbitPublisher(IConfiguration config)
        {
            var factory = new ConnectionFactory
            {
                HostName = config["RabbitMQ:Host"] ?? "rabbitmq",
                UserName = config["RabbitMQ:User"] ?? "guest",
                Password = config["RabbitMQ:Pass"] ?? "guest"
            };
            _conn = factory.CreateConnection();
            _channel = _conn.CreateModel();
            _channel.ExchangeDeclare(Exchange, ExchangeType.Topic, durable: true);
        }

        public void Publish(string routingKey, object message)
        {
            var body = Encoding.UTF8.GetBytes(JsonSerializer.Serialize(message));
            var props = _channel.CreateBasicProperties();
            props.Persistent = true;
            _channel.BasicPublish(Exchange, routingKey, props, body);
        }

        public void Dispose() { _channel?.Dispose(); _conn?.Dispose(); }
    }
}
