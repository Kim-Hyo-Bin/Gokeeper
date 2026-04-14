package api

import (
	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine, h *Handler) {
	r.GET("/health", Health)
	v1 := r.Group("/v1")
	{
		v1.POST("/licenses", h.IssueLicense)
		v1.POST("/licenses/verify", h.VerifyLicense)
		v1.GET("/licenses/:id", h.GetLicense)
		v1.POST("/licenses/:id/revoke", h.RevokeLicense)
	}
}
