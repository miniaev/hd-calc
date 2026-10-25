# cgo is on from the start: the Swiss Ephemeris binding needs it.
FROM golang:1.26-alpine AS build
RUN apk add --no-cache build-base
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/hd-calc ./cmd/hd-calc

FROM alpine:3.22
# tzdata is the source of historical time zones; its version goes into every response.
RUN apk add --no-cache tzdata \
 && echo "TZDB_VERSION=$(head -1 /usr/share/zoneinfo/tzdata.zi | awk '{print $3}')" > /etc/tzdb.env
WORKDIR /app
COPY --from=build /out/hd-calc /app/hd-calc
# COPY ephe/ /app/ephe/   # Swiss Ephemeris .se1 files — once vendored
EXPOSE 8080
CMD ["sh", "-c", ". /etc/tzdb.env && export TZDB_VERSION && exec /app/hd-calc"]
