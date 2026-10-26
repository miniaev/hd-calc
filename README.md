# hd-calc

Stateless HTTP service that calculates Human Design charts: birth time and
place in, chart JSON out — gates, lines, channels, centers, type, authority,
profile, definition, incarnation cross, transits.

Planetary positions come from the [Swiss Ephemeris](https://www.astro.com/swisseph/)
by Astrodienst AG, used under the AGPL. This service is therefore licensed
under the **GNU AGPL-3.0** — see [LICENSE](LICENSE).

No users, no database, no interpretations: only calculation. Output is keys
(`projector`, `20-34`, `solar_plexus`), not text.

## API

[`api/openapi.yaml`](api/openapi.yaml):

| Endpoint | What |
|---|---|
| `POST /v1/chart` | chart for a local birth time + IANA zone |
| `POST /v1/chart/range` | which chart elements change inside an uncertain birth-time window |
| `GET /v1/transits?at=` | transit activations at a moment |
| `POST /v1/transits/overlay` | channels and centers a transit completes in a given chart |
| `GET /v1/version` | engine version, ephemeris, tzdb, source URL |
| `GET /healthz` | liveness |

Every result carries `engine_version` = `<calc semver>/<ephemeris id>`.

## Layout

| Path | What |
|---|---|
| `internal/ephemeris` | `Ephemeris` interface — the only door to planetary positions |
| `internal/ephemeris/swisseph` | Swiss Ephemeris implementation (cgo) |
| `internal/hd` | longitudes → chart; knows only the interface |
| `internal/server` | HTTP |
| `testdata/golden` | reference charts every engine must reproduce |

## Run

```bash
go test ./...
go run ./cmd/hd-calc          # :8080
docker build -t hd-calc . && docker run --rm -p 8080:8080 hd-calc
```

## Status

All endpoints work (Swiss Ephemeris 2.10.03, DE441 files for 1800–2400, true lunar node).

`testdata/candidates.json` lists 26 charts picked to cover the golden checklist; each becomes
a golden fixture once cross-checked against public calculators.
