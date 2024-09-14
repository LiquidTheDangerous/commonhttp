package main

import (
	"io"
	"net/http"

	"github.com/LiquidTheDangerous/commonhttp/controller"
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

func main() {
	mux := &http.ServeMux{}
	controller.MustRegisterController(mux, &EchoController{})
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
