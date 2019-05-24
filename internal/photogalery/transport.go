package photogalery

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MakeHandler returns a handler for the photogalery service.
func MakeHandler(s Service) http.Handler {
	router := gin.Default()

	router.POST("/photos", makeUploadEndpoint(s))
	router.GET("/photos", makePhotosEndpoint(s))
	router.DELETE("/photos/:id", makeDeleteEndpoint(s))

	return router
}
