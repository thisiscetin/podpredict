Here’s the updated and final **`README.md`**, now with the Google Sheets data source clearly listed right at the top, along with the live demo and quick example usage.

---

# podpredict

![Tests](https://github.com/thisiscetin/podpredict/actions/workflows/test.yml/badge.svg)
[![Go Report](https://goreportcard.com/badge/github.com/thisiscetin/podpredict)](https://goreportcard.com/report/github.com/thisiscetin/podpredict)
[![Go Reference](https://pkg.go.dev/badge/github.com/thisiscetin/podpredict.svg)](https://pkg.go.dev/github.com/thisiscetin/podpredict)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

📊 **Live demo:** [https://podpredict.kinematiks.com/predictions](https://podpredict.kinematiks.com/predictions)
📄 **Data source:** [Google Sheet](https://docs.google.com/spreadsheets/d/1fMGuVN9FY5jWt_YHH0MPoIiriYW4AJjMaHoDW-2Jdm4)

You can try it instantly with:

```bash
curl -s -X POST https://podpredict.kinematiks.com/predict \
  -H 'Content-Type: application/json' \
  -d '{"gmv":1000000,"users":90000,"marketing_cost":150000}' | jq
```

---

**podpredict** is a lightweight Go service that predicts the number of **Front-End (FE)** and **Back-End (BE)** Kubernetes pods you’ll need based on daily KPIs:

* **GMV** (Gross Merchandise Value)
* **Users** (active user count)
* **Marketing Cost** (spend)

It trains two small linear regression models (FE + BE) and exposes a minimal HTTP API for prediction and storage.

---

## 🚀 Run with Docker

### 1️⃣ Pull or build the image

Pull from Docker Hub:

```bash
docker pull thisiscetin/podpredict:0.0.1
```

Or build locally:

```bash
docker build -t thisiscetin/podpredict:0.0.1 .
```

---

### 2️⃣ Run

Provide your Google Sheets credentials and spreadsheet ID:

```bash
docker run --rm -p 7000:7000 \
  -e GOOGLE_SHEETS_CREDENTIALS="$(cat creds.json)" \
  -e GOOGLE_SHEETS_SPREADSHEET_ID="1fMGuVN9FY5jWt_YHH0MPoIiriYW4AJjMaHoDW-2Jdm4" \
  thisiscetin/podpredict:0.0.1
```

Then open: [http://localhost:7000](http://localhost:7000)

---

## 🧩 API Endpoints

| Method | Path           | Description                 |
| :----: | :------------- | :-------------------------- |
|  `GET` | `/healthz`     | Health check                |
|  `GET` | `/predictions` | List all stored predictions |
| `POST` | `/predict`     | Predict FE/BE pods          |

Example request (local):

```bash
curl -s -X POST http://localhost:7000/predict \
  -H 'Content-Type: application/json' \
  -d '{"gmv":1000,"users":50,"marketing_cost":12}' | jq
```

Example response:

```json
{
  "id": "c9d9e0dc-4617-495e-a8c8-6ababead2571",
  "timestamp": "2025-01-02T12:34:56Z",
  "input": {"gmv":1000,"users":50,"marketing_cost":12},
  "fe_pods": 5,
  "be_pods": 3
}
```

---

## ⚙️ Configuration

| Env Variable                   | Description                                  |
| ------------------------------ | -------------------------------------------- |
| `GOOGLE_SHEETS_CREDENTIALS`    | Full JSON string of a Google service account |
| `GOOGLE_SHEETS_SPREADSHEET_ID` | Spreadsheet ID (from the URL)                |

### Example Sheet Layout

| Date       | GMV   | Users | MarketingCost | FEPods | BEPods |
| ---------- | ----- | ----- | ------------- | ------ | ------ |
| 01/01/2025 | 10000 | 50    | 100           | 3      | 2      |
| 02/01/2025 | 12000 | 70    | 150           |        |        |

Rows with both pods → used for **training**
Rows missing pods → **predicted** and stored at runtime

---

## 🧱 Architecture (Overview)

```
Google Sheets ──▶ Fetcher (gsheets)
      │
      ▼
  Regression Model (FE & BE)
      │
      ▼
  In-memory Store
      │
      ▼
   HTTP API (/predict, /predictions, /healthz)
```

---

## 🧪 Test locally

```bash
go test ./... -race
```

---

## 📜 License

MIT — see [LICENSE](LICENSE)

---

**🌐 Live version:** [https://podpredict.kinematiks.com/predictions](https://podpredict.kinematiks.com/predictions)
**📄 Data source:** [Google Sheet](https://docs.google.com/spreadsheets/d/1fMGuVN9FY5jWt_YHH0MPoIiriYW4AJjMaHoDW-2Jdm4)
**🐳 Docker image:** [`thisiscetin/podpredict:0.0.1`](https://hub.docker.com/r/thisiscetin/podpredict)
