package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marlonseben/cinema-booking/internal/reservas"
)

type Handler struct {
	service *reservas.Service
}

func NewHandler(service *reservas.Service) *Handler {
	return &Handler{service: service}
}

type reservarRequest struct {
	FilmeID   string `json:"filme_id" example:"matrix"`
	AssentoID string `json:"assento_id" example:"A1"`
	UsuarioID string `json:"usuario_id" example:"marlon"`
}

type erroResponse struct {
	Erro string `json:"erro"`
}

// Reservar godoc
//
//	@Summary		Reserva um assento
//	@Description	Confirma a reserva de um assento para um filme, de forma serializada por assento
//	@Tags			reservas
//	@Accept			json
//	@Produce		json
//	@Param			reserva	body		reservarRequest	true	"Dados da reserva"
//	@Success		201		{object}	reservas.Reserva
//	@Failure		400		{object}	erroResponse
//	@Failure		409		{object}	erroResponse	"Assento já ocupado"
//	@Failure		500		{object}	erroResponse
//	@Router			/reservas [post]
func (h *Handler) Reservar(c *gin.Context) {
	var req reservarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	r := reservas.Reserva{
		FilmeID:   req.FilmeID,
		AssentoID: req.AssentoID,
		UsuarioID: req.UsuarioID,
		Status:    "confirmada",
	}

	if err := h.service.Reservar(c.Request.Context(), r); err != nil {
		c.JSON(statusParaErro(err), gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, r)
}

// ListarReservas godoc
//
//	@Summary		Lista as reservas de um filme
//	@Description	Retorna todas as reservas confirmadas para o filme informado
//	@Tags			reservas
//	@Produce		json
//	@Param			filmeId	path		string	true	"ID do filme"
//	@Success		200		{array}		reservas.Reserva
//	@Router			/filmes/{filmeId}/reservas [get]
func (h *Handler) ListarReservas(c *gin.Context) {
	filmeID := c.Param("filmeId")

	reservas := h.service.ListarReservas(filmeID)
	c.JSON(http.StatusOK, reservas)
}

func statusParaErro(err error) int {
	switch {
	case errors.Is(err, reservas.ErrAssentoOcupado):
		return http.StatusConflict
	case errors.Is(err, reservas.ErrFilmeIDVazio),
		errors.Is(err, reservas.ErrAssentoIDVazio),
		errors.Is(err, reservas.ErrUsuarioIDVazio):
		return http.StatusBadRequest
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}
