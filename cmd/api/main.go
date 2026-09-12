package main

import (
	"log"

	"github.com/marlonseben/cinema-booking/internal/reservas"
	httptransport "github.com/marlonseben/cinema-booking/internal/transport/http"
)

func main() {
	store := reservas.NewMemoryStore()
	service := reservas.NewService(store)
	router := httptransport.NewRouter(service)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
