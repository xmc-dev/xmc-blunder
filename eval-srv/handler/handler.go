package handler

import (
	"github.com/xmc-dev/xmc/eval-srv/service"
)

var srv *service.Service

func InitHandler() {
	srv = service.MainService
}
