package main

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/LiquidTheDangerous/commonhttp/controller"
	"github.com/LiquidTheDangerous/commonhttp/middleware"
)

type ExampleControllerWithMiddleware struct {
}

func (e *ExampleControllerWithMiddleware) Routes() controller.Routes {
	return controller.Routes{controller.Route("GET", "/api/hello", e.SayHello)}
}

func (e *ExampleControllerWithMiddleware) SayHello(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, "Hello World")
	panic("whatever")
}

func PanicHandlingMiddleware(w http.ResponseWriter, r *http.Request, next http.Handler) {
	defer func() {
		v := recover()
		if v == nil {
			return
		}
		var errstring string
		switch u := v.(type) {
		case error:
			errstring = u.Error()
		case string:
			errstring = u
		}
		slog.Warn("panic recovered", slog.String("error", errstring))
	}()
	next.ServeHTTP(w, r)
}

func LoggingMiddlewareFunc(w http.ResponseWriter, r *http.Request, next http.Handler) {
	slog.Log(r.Context(), slog.LevelInfo, "handling request for: "+r.RequestURI)
	next.ServeHTTP(w, r)
}

func main() {
	mux := http.NewServeMux()
	controller.MustRegisterController(mux,
		&ExampleControllerWithMiddleware{},
		controller.WithMiddlewares(
			middleware.MiddlewareFunc(LoggingMiddlewareFunc),
			middleware.MiddlewareFunc(PanicHandlingMiddleware),
		),
	)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
