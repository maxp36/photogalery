package photogalery

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func makeUploadEndpoint(s Service) func(*gin.Context) {
	return func(c *gin.Context) {
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		photo, err := s.Upload(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"result": *photo,
		})
	}
}

func makePhotosEndpoint(s Service) func(*gin.Context) {
	return func(c *gin.Context) {
		photos := s.Photos()
		c.JSON(http.StatusOK, gin.H{
			"result": photos,
		})
	}
}

func makeDeleteEndpoint(s Service) func(*gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		photo, err := s.Delete(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"result": *photo,
		})
	}
}
