# Cinema Booking API

## O problema

Múltiplos usuários precisam selecionar assentos para uma sessão de cinema. Isso precisa ocorrer de forma concorrente, sem correr o risco de dois usuários agendarem o mesmo assento, ou do cinema vender mais ingressos do que o número de assentos disponíveis.

## Possíveis soluções

### Fila única (síncrono)

Uma possível solução seria ter somente uma bilheteria e uma única fila. Assim, o primeiro da fila compra seu ingresso/assento, depois o segundo, depois o terceiro etc., e a bilheteria sempre sabe que o assento "x" já está ocupado, não podendo vender para outro comprador. O problema é que esse approach não é muito escalável. Se tivermos muitos acessos ao mesmo tempo (por exemplo, o lançamento de um filme blockbuster), levaria muito tempo para que todos os usuários pudessem fazer o booking de seus assentos.

### Múltiplas filas (paralelo)

Uma solução melhor seria ter mais de uma bilheteria (e consequentemente mais de uma fila). Todavia, essa solução adiciona um requisito novo: controlar de forma paralela os assentos que já foram escolhidos/comprados. Isso é feito com um approach "pessimista", ou seja, se uma bilheteria está processando a venda do assento "x", qualquer outra bilheteria que tentar vender esse mesmo assento precisa esperar até que a primeira venda se resolva (locking).

### Por que não usar um approach otimista?

Um approach otimista, onde não existe trava, pode até fazer sentido, porém dois usuários podem selecionar o mesmo assento, e na hora do pagamento, aquele que for confirmar por último vai ter um erro, pois o assento mudou para "ocupado" enquanto ele pagava, tornando a UX ruim.

## Abordagem escolhida

Múltiplas bilheterias (múltiplas instâncias da API) com locking pessimista por assento: nenhuma delas guarda o estado de reserva sozinha, e a concessão de um assento só é confirmada quando se garante, de forma serializada, que nenhuma outra bilheteria está processando o mesmo assento ao mesmo tempo.

## Arquitetura

- **api1/api2/api3** (`cmd/api`) — instâncias HTTP (Gin) stateless, atrás de um gateway.
- **gateway** — Nginx fazendo load balancing entre as APIs e rate limiting.
- **worker** (`cmd/worker`) — instância única que processa os comandos de reserva de forma serializada, garantindo o locking pessimista, e persiste no Postgres.
- **notifier** (`cmd/notifier`) — consome a fila de notificações e simula o aviso ao usuário após a confirmação da reserva.
- **RabbitMQ** — RPC síncrono (API → worker, via direct reply-to) e fila assíncrona de notificações (worker → notifier).
- **Postgres** — estado durável das reservas.

## Como rodar o projeto

### Requisitos

- Docker e Docker Compose
- Go 1.27+ (opcional, só para rodar/testar fora de containers)

### Subindo tudo com Docker Compose

```
docker compose up --build
```

Isso sobe: Postgres (`5433`), RabbitMQ (`5672`, management UI em `15672`), três instâncias da API atrás do gateway Nginx (`8080`), o `worker` e o `notifier`. As migrations do Postgres são aplicadas automaticamente pela API/worker ao iniciar.

Serviços expostos:

- API (via gateway): `http://localhost:8080`
- Swagger: `http://localhost:8080/swagger/index.html`
- RabbitMQ management: `http://localhost:15672` (usuário/senha `guest`/`guest`)

Para acompanhar os logs de um serviço específico:

```
docker compose logs -f worker notifier
```

### Testando a API

- Coleção do [Bruno](https://www.usebruno.com/) em `bruno/` (ambiente `Local` já aponta para `http://localhost:8080`).
- Ou via curl, por exemplo:
  ```
  curl -X POST http://localhost:8080/reservas \
    -H "Content-Type: application/json" \
    -d '{"filmeId":"bacurau","assentoId":"A1","usuarioId":"user-1"}'
  ```

### Frontend estático

Uma página simples de demonstração fica em `static/`. Basta abrir `static/index.html` diretamente no navegador (o campo "baseUrl" já aponta para `http://localhost:8080`, o gateway).

### Rodando sem Docker (dev local)

Requer Postgres e RabbitMQ acessíveis localmente (ou os `docker compose up postgres rabbitmq`).

```
export POSTGRES_DSN="postgres://cinema:cinema@localhost:5433/cinema_booking?sslmode=disable"
export RABBITMQ_URL="amqp://guest:guest@localhost:5672/"

go run ./cmd/api       # instância da API em :8080
go run ./cmd/worker    # processa as reservas
go run ./cmd/notifier  # consome as notificações
```

### Testes e checagens

```
make check   # build + vet + test (equivalente a: go build ./..., go vet ./..., go test ./... -race)
```
