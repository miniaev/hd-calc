// Command hd-calc serves Human Design chart calculations over HTTP.
//
// Env:
//
//	ADDR          listen address (default :8080)
//	SE_EPHE_PATH  directory with Swiss Ephemeris .se1 files (default /app/ephe)
//	TZDB_VERSION  tzdata version baked into the image (reported in responses)
//	SOURCE_URL    public source of exactly this build (default: repo URL)
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/miniaev/hd-calc/internal/ephemeris/swisseph"
	"github.com/miniaev/hd-calc/internal/hd"
	"github.com/miniaev/hd-calc/internal/server"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	eph := swisseph.New(env("SE_EPHE_PATH", "/app/ephe"))
	eng := hd.NewEngine(eph, swisseph.Name+"-"+swisseph.Version, env("TZDB_VERSION", "unknown"))

	h := server.New(server.Config{
		Engine:           eng,
		EphemerisName:    swisseph.Name,
		EphemerisVersion: swisseph.Version,
		SourceURL:        env("SOURCE_URL", "https://github.com/miniaev/hd-calc"),
	})

	addr := env("ADDR", ":8080")
	srv := &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
	}
	log.Printf("hd-calc %s listening on %s", eng.EngineVersion, addr)
	log.Fatal(srv.ListenAndServe())
}
