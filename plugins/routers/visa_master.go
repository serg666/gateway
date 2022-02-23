package routers

import "log"

type Router interface {
	Route()
}

type Route struct {
	ChannelID int
	AccountID int
}

type VisaMasterRouter struct {
	route *Route
}

func (vmr *VisaMasterRouter) Route() {
	log.Print("some message")
}

func NewVisaMasterRouter(route *Route) Router {
	return &VisaMasterRouter{
		route: route,
	}
}
