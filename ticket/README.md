# Ticket Service

HTTP microservice for ticket creation and decision-tree based problem resolution.

## API

- `POST /tickets/` creates ticket and starts decision tree at `start`
- `GET /tickets/` lists tickets
- `GET /tickets/{id}` returns ticket
- `PATCH /tickets/{id}` updates ticket fields
- `DELETE /tickets/{id}` soft deletes ticket
- `GET /tickets/{id}/decision` returns current decision node
- `POST /tickets/{id}/decision` selects next node with `{"next_node_id":"product_problem"}`
- `GET /decision-tree/{nodeID}` returns any decision node
- `GET /health/live`, `GET /health/ready`, `GET /metrics`
- `GET /swagger/` opens Swagger UI
- `GET /swagger/openapi.yaml` returns OpenAPI spec

## Run

```sh
docker compose up --build
```

Service: `http://localhost:8080`
Prometheus: `http://localhost:9090`
Grafana: `http://localhost:3000` (`admin` / `admin`)

## Events

RabbitMQ exchange defaults to `warehouse` and uses event type as topic routing key. Published event types:

- `ticket.created`
- `ticket.updated`
- `ticket.deleted`
- `ticket.decision_step_selected`
