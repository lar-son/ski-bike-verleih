// Fahrrad-Verleih: Go-Backend (In-Memory) + HTML/JS-Frontend.
package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"time"

	"bikeverleih/internal/handler"
	"bikeverleih/internal/store"
)

//go:embed web
var webFS embed.FS

func main() {
	addr := flag.String("addr", ":8080", "Adresse, z.B. :8080")
	seed := flag.Bool("seed", true, "Demodaten laden")
	flag.Parse()

	st := store.NewMemoryStore()
	if *seed {
		if err := st.Seed(time.Now()); err != nil {
			log.Fatalf("demodaten: %v", err)
		}
	}

	static, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("frontend: %v", err)
	}
	h := handler.New(st)
	log.Printf("Fahrrad-Verleih läuft auf http://localhost%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, h.Routes(http.FileServer(http.FS(static)))))
}
