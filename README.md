# URL checker

HTTP service for checking URL availability.

## Requirements

- Docker

## Technology stack

- Go 1.25

## Running

```bash
docker compose up --build
```

The service is available at `http://localhost:8123`.

## API

Check URLs:

```bash
curl -X POST http://localhost:8123/checks \
  -H 'Content-Type: application/json' \
  -d '{
    "urls": [
      "https://element.id.lv",
      "https://videorip.app",
      "https://alexandr.ooo"
    ]
  }'
```

Invalid request:

```bash
curl -X POST http://localhost:8123/checks \
  -H 'Content-Type: application/json' \
  -d '{
    "urls": [
      "https://element.id.lv",
      "invalid-url",
      "https://videorip.app",
      "https://alexandr.ooo"
    ]
  }'
```

Response contains HTTP status, response time and error (if check failed).

```json
{
    "results": [
        {
            "url": "https://url",
            "status_code": int,
            "duration_ms": int
        },
        {
            "url": "invalid-url",
            "status_code": 0,
            "duration_ms": 0,
            "error": string
        }
    ]
}
```
