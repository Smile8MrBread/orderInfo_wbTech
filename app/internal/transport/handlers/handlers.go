// Handlers and starting server
package handlers

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"orderInfo/app/internal/models"
	"sync"
)

type Shower interface {
	ShowData(ctx context.Context, uid string) (models.Order, error)
}

type Casher interface {
	CashReturner(ctx context.Context) ([]models.Order, error)
}

type ServerAPI struct {
	dataShower Shower
	casher     Casher
	cash       map[string]models.Order
	mu         sync.Mutex
}

func NewServer(server Shower, cash Casher) ServerAPI {
	return ServerAPI{
		dataShower: server,
		casher:     cash,
		cash:       map[string]models.Order{},
		mu:         sync.Mutex{},
	}
}

func (s *ServerAPI) Start(r chi.Router, addr string) {
	orders, err := s.casher.CashReturner(context.Background())
	if err != nil {
		log.Println("Failed to get cash from bd")
	}

	for i := range orders {
		s.cash[orders[i].Uid] = orders[i]
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./client/index.html")
	})

	r.Get("/data/{uid}", func(w http.ResponseWriter, r *http.Request) {
		uid := chi.URLParam(r, "uid")

		if _, in := s.cash[uid]; !in {
			order, err := s.dataShower.ShowData(r.Context(), uid)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Internal error"))
				return
			}

			s.mu.Lock()
			s.cash[uid] = order
			s.mu.Unlock()
		}

		data, err := json.Marshal(s.cash[uid])
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Internal error"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(data)
	})

	http.ListenAndServe(addr, r)
}
