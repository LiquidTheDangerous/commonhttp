package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/LiquidTheDangerous/commonhttp/controller"
	"github.com/gorilla/mux"
)

type EchoController struct {
}

func (e *EchoController) ServeEcho(w http.ResponseWriter, r *http.Request) {
	_, err := io.Copy(w, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (e *EchoController) Routes() []controller.Route {
	// curl --request GET --url localhost:8080 --json '{"payload":"Hello, world!"}'
	return []controller.Route{{"GET", "/", e.ServeEcho, nil}}
}

func RegisterGorillaMux(m any, handler http.Handler, route *controller.Route) error {
	r, ok := m.(*mux.Router)
	if !ok {
		return fmt.Errorf("route %T is not a mux.Router", m)
	}
	r.Handle(route.Pattern, handler).Methods(route.Method)
	return nil
}

func main() {
	m := mux.NewRouter()
	controller.MustRegisterController(m, &EchoController{}, controller.WithHandlerRegistrarFunc(RegisterGorillaMux))
	if err := http.ListenAndServe(":8080", m); err != nil {
		panic(err)
	}
}
