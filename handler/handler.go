package handler

import "fm/service"

type handler struct {
	service service.Service
}

func New(s service.Service) *handler {
	return &handler{service: s}
}
