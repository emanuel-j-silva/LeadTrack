# Plataforma de Avaliação de Liderados (LeadTrack)

Plataforma web em que um líder avalia semanalmente os funcionários da sua hierarquia (subordinados diretos e indiretos). As respostas são notas inteiras de 1 a 4 com pesos por questão, são imutáveis após o envio e limitadas a uma avaliação por semana para cada par líder-avaliado.

---

## Sumário

1. [Stack](#stack)
2. [Arquitetura e Fluxo](#arquitetura-e-fluxo)
3. [Estrutura do repositório](#estrutura-do-repositório)
4. [Modelo de dados](#modelo-de-dados)
5. [Regras de negócio](#regras-de-negócio)
6. [Identificação do líder (sem login)](#identificação-do-líder-sem-login)
7. [Documentação da API e Endpoints](#documentação-da-api-e-endpoints)
8. [Setup e Execução (Instruções passo a passo)](#setup-e-execução-instruções-passo-a-passo)
9. [Convenções de código](#convenções-de-código)
10. [Testes](#testes)
11. [Premissas e decisões](#premissas-e-decisões)

---

## Stack

| Camada | Tecnologia |
|---|---|
| Backend | Go 1.23+, `chi` (router), `pgx/v5` (driver e pool), `golang-migrate` (migrations embutidas), `log/slog` (logs) |
| Banco | PostgreSQL 16 |
| Frontend | React + Vite + TypeScript, React Router, TanStack Query, react-hook-form + zod, Tailwind CSS |
| Infra | Docker + Docker Compose (db, api e web na mesma rede) |
| Docs de API | `docs/openapi.yaml` |
| Testes | `testing` + `testify` + `testcontainers-go` (backend), Vitest + Testing Library (frontend) |

## Arquitetura e Fluxo

```mermaid
flowchart LR
    U[Usuário] --> W[Frontend React :5173]
    W -- "HTTP JSON + header X-Employee-Id" --> A[API Go :8080]
    A --> DB[(PostgreSQL :5432)]

    subgraph API Go
        H[handler / http] --> S[service] --> R[repository / SQL]
    end
```

**Camadas do backend** (dependências apontam sempre para dentro):

- `http`: roteamento, middleware, decodificação e validação de payload, tradução de erros de domínio em status HTTP.
- `service`: casos de uso e regras de negócio (hierarquia, semana, cálculo de nota, visibilidade).
- `repository`: SQL parametrizado. Nenhuma regra de negócio.
- `domain`: entidades, erros de domínio e funções puras (início da semana, cálculo da nota).

O service define as interfaces de repositório de que precisa (interfaces pequenas, declaradas no consumidor), o que permite testar com fakes.

**Fluxo de uma avaliação**

```mermaid
sequenceDiagram
    participant UI
    participant API
    participant DB
    UI->>API: POST /api/employees/{id}/evaluations (X-Employee-Id: líder)
    API->>DB: Verifica se avaliado está na subárvore (CTE recursiva)
    API->>API: Valida 6 respostas (1-4), calcula week_start e nota
    API->>DB: Executa transação (BEGIN, INSERT evaluation e respostas, COMMIT)
    DB-->>API: Retorna sucesso ou violação de UNIQUE (23505)
    API-->>UI: Retorna 201 Created ou 409 Conflict
```

## Estrutura do repositório

```
.
├── backend/
│   ├── cmd/server/main.go          # composição: config, pool, migrations, router
│   ├── internal/
│   │   ├── config/                 # variáveis de ambiente
│   │   ├── domain/                 # entidades, erros, week.go, score.go
│   │   ├── http/                   # router, middleware, handlers, dto, errors
│   │   ├── service/                # casos de uso
│   │   └── repository/             # SQL (pgx)
│   ├── migrations/                 # 000001_schema.up.sql, 000002_seed.up.sql ... (embed)
│   ├── Dockerfile
│   └── go.mod
├── frontend/                       # React + Vite + TS
├── docker-compose.yml
├── Makefile
├── .env.example
└── README.md
```

## Modelo de dados

As tabelas `employee` e `leader_lead` vêm do dump fornecido. A relação `leader_lead` é **N:N**: um funcionário pode ter mais de um líder, então a hierarquia é um grafo e o código não pode assumir árvore.

```sql
-- Dump fornecido
CREATE TABLE employee (
    id            SERIAL       PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    email         VARCHAR(150) NOT NULL UNIQUE,
    position_name VARCHAR(100) NOT NULL
);

CREATE TABLE leader_lead (
    leader_id INT NOT NULL REFERENCES employee(id) ON DELETE CASCADE ON UPDATE CASCADE,
    lead_id   INT NOT NULL REFERENCES employee(id) ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY (leader_id, lead_id),
    CONSTRAINT chk_no_self_lead CHECK (leader_id <> lead_id)
);
CREATE INDEX idx_leader_lead_lead ON leader_lead(lead_id);

-- Novas
CREATE TABLE question (
    id     SERIAL       PRIMARY KEY,
    label  VARCHAR(120) NOT NULL UNIQUE,
    weight INT          NOT NULL CHECK (weight > 0)
);

CREATE TABLE evaluation (
    id           SERIAL PRIMARY KEY,
    evaluator_id INT NOT NULL REFERENCES employee(id),
    evaluated_id INT NOT NULL REFERENCES employee(id),
    week_start   DATE NOT NULL,             -- segunda-feira da semana, calculada no servidor
    score        NUMERIC(4,2) NOT NULL,     -- nota ponderada (1.00 a 4.00), gravada no envio
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (evaluator_id <> evaluated_id),
    UNIQUE (evaluator_id, evaluated_id, week_start)
);
CREATE INDEX idx_evaluation_evaluated ON evaluation(evaluated_id, created_at DESC);

CREATE TABLE evaluation_answer (
    evaluation_id INT NOT NULL REFERENCES evaluation(id),
    question_id   INT NOT NULL REFERENCES question(id),
    value         SMALLINT NOT NULL CHECK (value BETWEEN 1 AND 4),
    PRIMARY KEY (evaluation_id, question_id)
);
```

**Seed de perguntas**

| Questão | Peso |
|---|---|
| Entrega de Resultados | 25 |
| Execução e Qualidade do Trabalho | 20 |
| Capacidade de Aprendizado e Desenvolvimento | 20 |
| Resolução de Problemas e Pensamento Crítico | 15 |
| Colaboração, Influência e Liderança | 10 |
| Visão Estratégica e Potencial de Crescimento | 10 |

**Seed de funcionários:** os 20 funcionários e as relações do dump fornecido (sem os `DROP TABLE`). Hierarquia resultante:

```
Alice (CEO)
├─ Bob (CTO)
│  ├─ David (Eng Manager)
│  │  ├─ Henry (Sr Engineer) ── James, Karen
│  │  └─ Liam
│  ├─ Eva (Eng Manager) ── Isabelle, Mia, Noah
│  ├─ Grace, Quinn, Paul
├─ Carol (CFO) ── Rachel, Samuel
├─ Frank (PM) ── Olivia
└─ Tina (HR)
```

**Subordinados diretos e indiretos** (`UNION` e não `UNION ALL`, para deduplicar e evitar loop infinito em caso de ciclo):

```sql
WITH RECURSIVE subs(id, depth) AS (
    SELECT lead_id, 1 FROM leader_lead WHERE leader_id = $1
    UNION
    SELECT ll.lead_id, s.depth + 1
    FROM leader_lead ll JOIN subs s ON ll.leader_id = s.id
)
SELECT id, MIN(depth) AS depth FROM subs GROUP BY id;
```

## Regras de negócio

| # | Regra | Onde é garantida |
|---|---|---|
| R1 | Respostas são inteiros de 1 a 4, uma por questão (as 6 são obrigatórias) | Handler/service (`422`) + `CHECK` no banco |
| R2 | Nota = `Σ(valor × peso) / Σ(pesos)`, resultado entre 1.00 e 4.00 | `domain/score.go` (função pura, testada) |
| R3 | Uma avaliação por semana por par líder-avaliado | `UNIQUE (evaluator_id, evaluated_id, week_start)`. Erro `23505` vira `409` |
| R4 | Semana = segunda a domingo no fuso `APP_TIMEZONE` (padrão `America/Sao_Paulo`); o servidor calcula, o cliente não envia | `domain/week.go` |
| R5 | Avaliações são imutáveis | Não existem `PUT`, `PATCH` ou `DELETE` de avaliações |
| R6 | O líder só avalia quem está na sua subárvore (diretos e indiretos), mesmo que um subordinado já tenha avaliado o mesmo liderado na semana | Service (`403`) |
| R7 | Ninguém avalia a si mesmo | `CHECK` + service (`403`) |
| R8 | O usuário nunca vê a própria avaliação, nem de pares ou superiores. Só vê avaliações de subordinados, feitas por ele ou por subordinados dele | Toda query de leitura filtra pela subárvore do usuário atual |

Rotas de leitura fora da visibilidade respondem `404` (não vazam a existência do recurso) ou `403`; escolha uma e mantenha consistente (padrão deste projeto: `403`).

## Identificação do líder (sem login)

Não há login. O líder atual é identificado assim:

1. O frontend guarda o `employeeId` escolhido em `localStorage` (`currentEmployeeId`).
2. Um seletor no topo da UI lista todos os funcionários (`GET /api/employees`) e permite trocar de líder; a troca invalida o cache do TanStack Query.
3. Toda requisição leva o header `X-Employee-Id: <id>`.
4. O middleware do Go lê o header, confirma que o funcionário existe e o coloca no `context`. Sem header ou id inexistente: `401`.

**Isso é deliberadamente inseguro** (qualquer cliente pode se passar por qualquer funcionário). Em produção seria substituído por JWT ou sessão com cookie `HttpOnly`; apenas o middleware precisaria mudar.

`GET /api/employees` e `GET /health` não exigem o header.

## Documentação da API e Endpoints

Base: `/api`. JSON em UTF-8. Erros seguem o formato:

```json
{ "error": { "code": "ALREADY_EVALUATED_THIS_WEEK", "message": "..." } }
```

| Método | Rota | Descrição |
|---|---|---|
| GET | `/health` | Liveness |
| GET | `/api/employees` | Lista todos os funcionários (seletor de usuário) |
| GET | `/api/me` | Funcionário atual |
| GET | `/api/questions` | Perguntas e pesos |
| GET | `/api/subordinates` | Subordinados (diretos e indiretos) com profundidade, última avaliação visível e `canEvaluateThisWeek` |
| GET | `/api/employees/{id}/evaluations` | Histórico de avaliações visíveis do subordinado, mais recentes primeiro |
| GET | `/api/employees/{id}/evaluations/latest` | Avaliação visível mais recente, com respostas |
| POST | `/api/employees/{id}/evaluations` | Cria avaliação |

**POST body**

```json
{ "answers": [ { "questionId": 1, "value": 3 }, { "questionId": 2, "value": 4 } ] }
```

(devem estar presentes as 6 questões)

**Status HTTP**

| Código | Quando |
|---|---|
| 201 | Avaliação criada |
| 400 | JSON inválido ou `{id}` inválido |
| 401 | Header `X-Employee-Id` ausente ou inexistente |
| 403 | Avaliado fora da hierarquia do usuário, ou autoavaliação |
| 409 | Par já avaliado na semana corrente |
| 422 | Respostas incompletas, duplicadas, de questão inexistente ou fora de 1 a 4 |
| 500 | Erro inesperado (sem vazar detalhes) |

## Setup e Execução (Instruções passo a passo)

**Pré-requisitos:** Docker e Docker Compose. Para desenvolvimento local fora do container: Go 1.23+ e Node 20+.

### 1. Configuração de Variáveis de Ambiente
Copie o arquivo de exemplo para criar o seu `.env`:
```bash
cp .env.example .env
```

### 2. Execução via Docker Compose (Recomendado)
Para subir toda a stack (`db`, `api` e `web`) em uma única rede:
```bash
docker compose up --build
```
- **Frontend (Web):** `http://localhost:5173`
- **API Backend:** `http://localhost:8080` (Liveness em `GET /health`)

As migrations (schema e sementes) rodam automaticamente na subida da API. Para recriar o banco do zero: `docker compose down -v`.

### 3. Desenvolvimento local (Sem Docker para API/Web)
Se preferir rodar a API e o Frontend nativamente:
```bash
# Sobe apenas o banco de dados
docker compose up -d db

# Roda a API backend
cd backend && DATABASE_URL="postgres://app:app@localhost:5432/evaluation?sslmode=disable" go run ./cmd/server

# Em outro terminal, roda o frontend
cd frontend && npm install && npm run dev
```

### 4. Makefile
Comandos auxiliares disponíveis via Makefile:
```
make up        # docker compose up --build
make down      # docker compose down -v
make test      # executa testes do backend e frontend
make lint      # executa lints (go vet, eslint, etc.)
make fmt       # formata código (gofmt, prettier)
```

## Convenções de código

- **Go:** `gofmt`, `go vet` limpos. `context.Context` sempre como primeiro parâmetro. Erros embrulhados com `fmt.Errorf("...: %w", err)` e erros de domínio mapeados para status HTTP via `errors.Is`.
- **SQL:** Sempre parametrizado (`$1`, `$2`), prevendo segurança e evitando N+1.
- **Transações:** Criação de avaliação e respostas atômica em transação.
- **Frontend:** TypeScript estrito, sem `any`. TanStack Query para gerenciamento de estado do servidor e react-hook-form + zod para formulários.
- **Commits:** Conventional Commits (`feat:`, `fix:`, `docs:`, `test:`, `chore:`).

## Testes

| Nível | O que cobrir |
|---|---|
| Unitário (`domain`) | `weekStart` (virada de semana, fuso), cálculo da nota ponderada |
| Unitário (`service`) | Validação de respostas, hierarquia, autoavaliação, mapeamento de erros |
| Integração (`repository`) | Postgres real: CTE recursiva, UNIQUE por semana, isolamento de hierarquia |
| HTTP | Respostas por status code (201, 400, 401, 403, 409, 422) |
| Frontend | Formulário de avaliação e restrições de envio por semana |

## Premissas e decisões

1. **Hierarquia N:N.** `leader_lead` permite múltiplos líderes por funcionário; a subárvore é o fecho transitivo (CTE recursiva com `UNION`).
2. **Visibilidade (R8).** O usuário só enxerga avaliações cujo avaliado está na sua subárvore **e** cujo avaliador é ele mesmo ou alguém da sua subárvore.
3. **Semana.** Começa na segunda-feira no fuso `APP_TIMEZONE` configurado no servidor.
4. **Nota.** Média ponderada entre 1.00 e 4.00 baseada nos pesos cadastrados na tabela `question`.
5. **Autenticação.** Simulada por cabeçalho `X-Employee-Id` para fins de simulação e testes.
