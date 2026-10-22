# explicAI

API em Go que recebe um áudio, transcreve e devolve um resumo, usando APIs externas de LLM.

Projeto final do curso **Golang Completo: Fundamentos, APIs Web, Concorrência e LLMs** (Udemy).

## Como funciona

1. `POST /upload` valida o arquivo, cria o registro no PostgreSQL e responde na hora com status e progresso.
2. Em segundo plano (goroutine), o áudio é transcrito.
3. O resumo estruturado e o texto organizado rodam em paralelo (`errgroup`).
4. O status final é salvo. Se uma etapa falha, o registro guarda qual foi.

| Status               | Progresso |
|----------------------|-----------|
| `RECEIVED_FILE`      | 33        |
| `TRANSCRIBED`        | 66        |
| `SUMMARIZED`         | 100       |
| `TRANSCRIBED_FAILED` | 33        |
| `SUMMARIZED_FAILED`  | 66        |

## Estrutura

```
cmd/                    ponto de entrada
configuration/          config (viper) e montagem dos clients
internal/
  application/service/  regras do fluxo de resumo
  gateway/              interfaces: transcrição, resumo e repositório
  infrastructure/
    api/                controller (Echo)
    clients/            clients HTTP das APIs de IA (resty)
    db/                 acesso ao PostgreSQL (pgx)
```

O serviço depende só das interfaces em `gateway/`, então os serviços externos e o banco podem ser trocados por mocks nos testes.

## Rodando

Requisitos: Go, Docker e uma chave de API da OpenAI.

```bash
cp .env.example .env        # preencha OPENAI_API_KEY
docker compose up -d db     # PostgreSQL com o ddl.sql aplicado
go run ./cmd
```

A API sobe em `http://localhost:8080`.

### Variáveis

| Variável         | Padrão                |
|------------------|-----------------------|
| `OPENAI_API_KEY` | obrigatória           |
| `HOST`, `PORT`   | `0.0.0.0`, `8080`     |

A URL do banco, os modelos e os timeouts têm padrão em `configuration/configuration.go`: `postgres://admin:admin@localhost:5432/explicai`, `whisper-1` para transcrição e `gpt-4o` para resumo.

## Endpoints

| Método | Rota                      | Descrição                          |
|--------|---------------------------|------------------------------------|
| POST   | `/upload`                 | envia o áudio (campo `file`)       |
| GET    | `/summaries`              | lista os resumos                   |
| GET    | `/summaries/:externalId`  | detalhe com resumos e texto completo |
| DELETE | `/summaries/:externalId`  | remove um resumo                   |

Formatos aceitos: `.mp3`, `.mp4`, `.mpeg`, `.mpga`, `.m4a`, `.wav`, `.webm`.

```bash
curl -F "file=@audio.mp3" http://localhost:8080/upload
curl http://localhost:8080/summaries/<externalId>
```

Resposta do detalhe (campos `title`, `description`, `briefResume`, `mediumResume` e `fullText` aparecem quando o resumo termina):

```json
{
  "externalId": "…",
  "status": "SUMMARIZED",
  "progress": 100,
  "title": "…",
  "briefResume": "…",
  "mediumResume": "…",
  "fullText": "…"
}
```

## Testes

```bash
go test ./...                                   # todos
go test -coverprofile=coverage.out ./... -race  # com cobertura
go tool cover -html=coverage.out                # relatório no navegador
```

Os testes de banco usam PostgreSQL em container (testcontainers), então o Docker precisa estar rodando. Os clients de IA são testados com respostas simuladas (sucesso, vazio, sem `choices`, JSON inválido).

## Stack

Go, Echo, resty, pgx, PostgreSQL, viper, Zap, errgroup, testify, mockery, testcontainers.

`front.html` é só uma página para testar a API no navegador.
