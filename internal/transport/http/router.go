package http

import (
	"github.com/gin-gonic/gin"
	"github.com/marlonseben/cinema-booking/internal/reservas"
)

func NewRouter(service *reservas.Service) *gin.Engine {
	router := gin.Default()
	handler := NewHandler(service)

	router.POST("/reservas", handler.Reservar)
	router.GET("/filmes/:filmeId/reservas", handler.ListarReservas)

	return router
}
