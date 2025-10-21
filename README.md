# podpredict

![Tests](https://github.com/thisiscetin/podpredict/actions/workflows/test.yml/badge.svg)
[![Go Report](https://goreportcard.com/badge/github.com/thisiscetin/podpredict)](https://goreportcard.com/report/github.com/thisiscetin/podpredict)
[![Go Reference](https://pkg.go.dev/badge/github.com/thisiscetin/podpredict.svg)](https://pkg.go.dev/github.com/thisiscetin/podpredict)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**podpredict** is a small Go service that **predicts the number of Front‑End (FE) and Back‑End (BE) Kubernetes pods** you’ll need based on daily business KPIs:

* **GMV** (Gross Merchandise Value)
* **Users** (count)
* **MarketingCost** (spend)

Under the hood, it trains **two separate multivariate linear regression models** (one for FE, one for BE) and exposes a minimal **HTTP API** to predict and persist results. A **Google Sheets fetcher** is included for training data, and an **in‑memory store** records predictions.

---

## Table of contents

* [High‑level architecture](#high-level-architecture)
* [API quickstart](#api-quickstart)
* [Running locally](#running-locally)
    * [With Go](#with-go)
    * [With Docker](#with-docker)
* [Training data sources](#training-data-sources)
    * [Google Sheets format](#google-sheets-format)
* [Model details](#model-details)
* [Project layout](#project-layout)
* [Configuration](#configuration)
* [Validation & error handling](#validation--error-handling)
* [Testing & CI](#testing--ci)
* [Extending the project](#extending-the-project)
* [License](#license)

---

## High‑level architecture

```
+-----------------+          +-------------------+          +---------------------------+
| Google Sheets   |  Fetch   | Fetcher (gsheets) |  Metrics | Model (linreg: FE + BE)   |
|                 +--------->+ or mock fetcher   +--------->+                           |
+-----------------+          +-------------------+    KPIs  +------------+--------------+
                                                     (GMV, Users, MC)   |
                                                                        |
                                          +-----------------------------v----------------------------+
                                          |  HTTP API (net/http, ServeMux)                           |
                                          |  /predict, /predictions, /healthz                        |
                                          +-----------------------------+----------------------------+
                                                                        |
                                                                        v
                                                            +-----------------------+
                                                            | Store (in‑memory)     |
                                                            | []Prediction          |
                                                            +-----------------------+
```

* **Fetcher** (`internal/fetcher`): pluggable data source; `gsheets` implementation fetches historical daily rows.
* **Model** (`internal/model/linreg`): two linear regressions trained on GMV, Users, MarketingCost.
* **API** (`internal/api`): endpoints to predict, list predictions, and health checks.
* **Store** (`internal/store/inmemory`): thread‑safe in‑memory slice protected by RWMutex; returns copies to avoid aliasing bugs.
* **Server** (`cmd/server`): wires everything, performs initial training, and serves on `:8080`.

---

## API quickstart

By default the server listens on **`http://localhost:8080`**.

### `POST /predict`

**Request body** (JSON):

```json
{
  "gmv": 13224723.00,
  "users": 123,
  "marketing_cost": 1000.5
}
```

**Response** `201 Created`:

```json
{
  "id":"c9d9e0dc-4617-495e-a8c8-6ababead2571",
  "timestamp": "2025-01-02T12:34:56Z",
  "input": {
    "gmv": 13224723,
    "users": 123,
    "marketing_cost": 1000.5
  },
  "fe_pods": 7,
  "be_pods": 3
}
```

Failure cases:

* `400 Bad Request` – invalid JSON body.
* `500 Internal Server Error` – model prediction error or store append error.

### `GET /predictions`

Returns all predictions stored so far (most recent first/last depending on the store’s append order).

**Response** `200 OK`:

```json
[
  {
    "id":"c9d9e0dc-4617-495e-a8c8-6ababead2571",
    "timestamp": "2025-01-02T12:34:56Z",
    "input": { 
      "gmv": 10, 
      "users": 1, 
      "marketing_cost": 1
    },
    "fe_pods": 5,
    "be_pods": 4
  }
]
```

Failure cases:

* `500 Internal Server Error` – store listing error.

### `GET /healthz`

Simple liveness & dependency check.

**Response** `200 OK`:

```json
{
  "status": "ok" | "degraded",
  "store_ok": true,
  "model_ok": true,
  "timestamp": "2025-01-02T12:34:56Z"
}
```

* `status: "degraded"` when the store cannot be listed.
* `405 Method Not Allowed` if a wrong method is used on any endpoint.

---

## Running locally

### With Go

> Requires **Go 1.25+** (per `go.mod`).

```bash
# 1) Run tests to sanity check your toolchain
go test ./... -race

# 2) Start the server (uses mock fetcher in the sample main)
go run ./cmd/server
# listening on :8080
```

Then, in another shell:

```bash
curl -s http://localhost:8080/healthz | jq
curl -s -X POST http://localhost:8080/predict \
  -H 'Content-Type: application/json' \
  -d '{"gmv":1000,"users":50,"marketing_cost":12}' | jq
curl -s http://localhost:8080/predictions | jq
```

### With Docker

A multi‑stage `Dockerfile` is included.

```bash
docker build -t podpredict:local .
docker run --rm -p 8080:8080 podpredict:local
```

> The image defaults to a non‑root user and exposes port `8080`.

---

## Training data sources

At server start, the API constructor fetches data and **trains the model**:

```go
// internal/api/handler.go:
data, err := f.Fetch()
if err != nil { /* ... */ }
if err := m.Train(data); err != nil { /* ... */ }
```

Two fetchers are provided:

1. **Mock fetcher** (`internal/fetcher/mock`): used in tests and in the sample `cmd/server/main.go`.
2. **Google Sheets fetcher** (`internal/fetcher/gsheets`): production‑like, authenticated via service account.

### Google Sheets format

The gsheets fetcher reads **`Sheet1!A:F`** with **dd/mm/yyyy** dates:

| Column | Field          | Type    | Notes                                   |
| -----: | -------------- | ------- | --------------------------------------- |
|      A | Date           | string  | `dd/mm/yyyy` (parsed with `02/01/2006`) |
|      B | GMV            | string  | may include thousands separators `,`    |
|      C | Users          | string  | integer                                 |
|      D | Marketing Cost | string  | float (commas stripped)                 |
|      E | FE Pods        | string? | optional; used for training if present  |
|      F | BE Pods        | string? | optional; used for training if present  |

* Rows with both pods present are used for **training**.
* Rows missing pods are **predicted** after training (see the sample main: it appends predictions for such rows to the in‑memory store).

**Environment variables** expected by the wired (commented) real fetcher in `cmd/server/main.go`:

* `GOOGLE_SHEETS_SPREADSHEET_ID` – the spreadsheet ID.
* `GOOGLE_SHEETS_CREDENTIALS` – the **entire JSON** of a service account (string), which must have read access to the sheet.

> In Google Sheets UI, share the sheet with the service account email.

---

## Model details

Package: `internal/model/linreg`
Library: [`github.com/sajari/regression`](https://pkg.go.dev/github.com/sajari/regression)

* **Features** (`internal/model.Features`):

  ```go
  type Features struct {
    GMV           float64 `json:"gmv"`
    Users         float64 `json:"users"`
    MarketingCost float64 `json:"marketing_cost"`
  }
  ```
* The `metrics.Daily` domain type exposes:

    * `Features() []float64 { GMV, Users, MarketingCost }`
    * `Pods() (fe int, be int, ok bool)` and `HasPods()` helpers.
* **Two models** run in parallel (FE and BE). Each is a regression with:

    * One intercept term.
    * Three numeric features: GMV, Users, MarketingCost.
* **Prediction post‑processing**:

    * `safeRound(v)` → rounds to nearest int and **defends against NaN/Inf** by falling back to `1`.
    * `clampMinInt(v, 1)` → **never returns less than 1** pod.
* The API constructor (`New`) **trains at startup**. You can retrain later by calling the handler’s `Retrain(ctx)` (provided for future endpoints/CLI).

---

## Project layout

```
.
├── cmd/
│   └── server/                   # Wires fetcher+model+store; starts HTTP server on :8080
│       └── main.go
├── internal/
│   ├── api/                      # HTTP layer (handlers, routes, helper types)
│   │   ├── handler.go
│   │   ├── handler_test.go
│   │   ├── routes.go
│   │   └── types.go
│   ├── fetcher/
│   │   ├── fetcher.go            # Fetcher interface
│   │   ├── gsheets/              # Google Sheets implementation (+ tests)
│   │   │   ├── fetcher.go
│   │   │   └── fetcher_test.go
│   │   └── mock/                 # Test/mock fetcher
│   │       └── mock_fetcher.go
│   ├── metrics/                  # Domain model + validation (Daily)
│   │   ├── metrics.go
│   │   └── metrics_test.go
│   ├── model/
│   │   ├── model.go              # Model interface + Features
│   │   ├── linreg/               # Linear regression implementation (+ tests)
│   │   │   ├── model.go
│   │   │   └── model_test.go
│   │   └── mock/                 # Mock model for API tests
│   │       └── mock_model.go
│   └── store/
│       ├── store.go              # Store interface + Prediction type
│       └── inmemory/             # Thread-safe in-memory store (+ tests)
│           ├── store.go
│           └── store_test.go
├── .github/workflows/test.yml    # CI: go test ./internal/... with race+coverage
├── Dockerfile
├── go.mod / go.sum
├── LICENSE (MIT)
└── README.md
```

---

## Configuration

| Setting                        | Where                            | Description                                     |
| ------------------------------ | -------------------------------- | ----------------------------------------------- |
| `:8080` listen address         | `cmd/server/main.go`             | Hardcoded in sample main. Adjust to your needs. |
| `GOOGLE_SHEETS_SPREADSHEET_ID` | env (used in commented wiring)   | Spreadsheet ID to read.                         |
| `GOOGLE_SHEETS_CREDENTIALS`    | env (used in commented wiring)   | Service account JSON (full string).             |
| Sheets range                   | `internal/fetcher/gsheets`       | `Sheet1!A:F` by default.                        |
| Date format                    | `internal/fetcher/gsheets`       | `02/01/2006` (dd/mm/yyyy).                      |
| API request timeout            | `internal/api.New(..., timeout)` | Defaulted to 5s if non‑positive.                |

---

## Validation & error handling

**Domain validation** (`internal/metrics`):

* `NewDaily(...)` returns well‑typed rows or errors:

    * `ErrInvalidDate` for zero date,
    * `ErrGmvNegative`, `ErrUsersNegative`, `ErrMarketingCostNegative` for negative values.
* `Features()` is always `[GMV, Users, MarketingCost]`.

**Google Sheets parsing** (`internal/fetcher/gsheets`):

* Commas in numeric strings are **removed** before parsing floats.
* `FE Pods` / `BE Pods` are **optional**; parse errors are logged and ignored.

**HTTP layer** (`internal/api`):

* Enforces HTTP methods (405 on mismatch).
* `POST /predict`:

    * `400` for bad JSON.
    * `500` for model failure or store append failure.
    * Returns the persisted `store.Prediction` with a server‑side timestamp.
* `GET /predictions`:

    * Returns a **copy** of internal slice (store returns copies to avoid aliasing).
* `GET /healthz`:

    * Returns `"ok"` or `"degraded"`; checks store by calling `List`.

---

## Testing & CI

**Unit tests** cover:

* `metrics` validation and helpers.
* `linreg` model (including coefficient sanity with synthetic data).
* `gsheets` parser (`parseRow`) for valid/invalid shapes.
* `inmemory` store (thread safety, copy‑on‑list).
* `api` handlers (successful flow, invalid JSON, model error, store error, listing, etc.).

Run them locally:

```bash
go test ./... -race
# or (as in CI)
go test ./internal/... -v -race -covermode=atomic -coverprofile=coverage.out
go tool cover -func=coverage.out
# HTML coverage:
go tool cover -html=coverage.out
```

CI: GitHub Actions workflow in `.github/workflows/test.yml` runs tests with race detector and coverage.

---

## Extending the project

* **Persistent storage**: Replace `internal/store/inmemory` with a Postgres or Redis store. The interface is intentionally small:

  ```go
  type Store interface {
    Append(ctx context.Context, r Prediction) error
    List(ctx context.Context) ([]Prediction, error)
  }
  ```
* **More features**: Add signals like latency, cache hit ratio, traffic mix, etc., to `metrics.Daily` and plumb them through `Features`.
* **Alternative models**: Swap `linreg` with a tree/ensemble model (or even a small neural net), behind the `model.Model` interface.
* **Periodic retraining**: add a background job or expose a `/retrain` endpoint that calls `Handler.Retrain(ctx)`.
* **Configurable port**: read `PORT` from env in `cmd/server/main.go`.

---

## License

This project is licensed under the **MIT License** — see [LICENSE](LICENSE).

---

### Appendix A — Example `curl` session

```bash
# Health
curl -s http://localhost:8080/healthz | jq

# Predict
curl -s -X POST http://localhost:8080/predict \
  -H 'Content-Type: application/json' \
  -d '{"gmv":1000,"users":50,"marketing_cost":12}' | jq

# List predictions
curl -s http://localhost:8080/predictions | jq
```

### Appendix B — Minimal Google Sheets wiring (for production)

In `cmd/server/main.go` (the snapshot shows this wiring commented out), you would roughly do:

```go
import (
    "context"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/thisiscetin/podpredict/internal/api"
    "github.com/thisiscetin/podpredict/internal/fetcher/gsheets"
    "github.com/thisiscetin/podpredict/internal/model/linreg"
    "github.com/thisiscetin/podpredict/internal/store/inmemory"
)

func main() {
    ctx := context.Background()

    sheetID := os.Getenv("GOOGLE_SHEETS_SPREADSHEET_ID")
    creds   := os.Getenv("GOOGLE_SHEETS_CREDENTIALS")
    if sheetID == "" || creds == "" { log.Fatal("missing gsheets env") }

    f, err := gsheets.NewFetcher(ctx, []byte(creds), sheetID)
    if err != nil { log.Fatal(err) }

    h, err := api.New(linreg.NewModel(), f, inmemory.NewStore(), 5*time.Second)
    if err != nil { log.Fatal(err) }

    log.Fatal(http.ListenAndServe(":8080", api.Routes(h)))
}
```