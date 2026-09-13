package http

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	_ "github.com/marlonseben/cinema-booking/docs"
	"github.com/marlonseben/cinema-booking/internal/reservas"
)

func NewRouter(service *reservas.Service) *gin.Engine {
	router := gin.Default()
	handler := NewHandler(service)

	router.POST("/reservas", handler.Reservar)
	router.GET("/filmes/:filmeId/reservas", handler.ListarReservas)
	router.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	return router
}
