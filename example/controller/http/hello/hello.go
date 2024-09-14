package main

import (
	"io"
	"net/http"

	"github.com/LiquidTheDangerous/commonhttp/controller"
)

type HelloController struct {
}

func (h *HelloController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		name = "anonymous"
	}
	_, _ = io.WriteString(w, "Hello, "+name+"!\n")
}

func (h *HelloController) Routes() controller.Routes {
	// curl --request GET --url localhost:8080/hello?name=YourName
	return controller.Routes{controller.Route("GET", "/hello", h)}
}

func main() {
	mux := http.NewServeMux()
	controller.MustRegisterController(mux, &HelloController{})
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
