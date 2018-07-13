package errors

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/errors"
	"github.com/sirupsen/logrus"
)

func err(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"error": msg})
}

func BadRequest(c *gin.Context) {
	err(c, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
}

func InternalServerError(c *gin.Context) {
	err(c, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
}

func NotFound(c *gin.Context) {
	err(c, http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

func Unauthorized(c *gin.Context) {
	c.Writer.Header().Set("WWW-Authenticate", "Bearer realm=\"XMC\"")
	err(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
}

func Error(c *gin.Context, me *errors.Error, e error) {
	if me.Code == http.StatusInternalServerError {
		c.Error(e)
		logrus.WithField("prefix", "httperror").WithField("e", fmt.Sprintf("%+v", e)).Error("Internal errors")
		InternalServerError(c)
	} else {
		err(c, int(me.Code), me.Detail)
	}
}
