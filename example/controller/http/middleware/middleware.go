package main

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/LiquidTheDangerous/commonhttp/controller"
	"github.com/LiquidTheDangerous/commonhttp/middleware"
)

type ExampleControllerWithMiddleware struct {
}

func (e *ExampleControllerWithMiddleware) Routes() controller.Routes {
	return controller.Routes{
		controller.Route("GET", "/api/hello", e.SayHello),
		controller.Route("POST", "/api/hi", e.SayHi),
	}
}

func (e *ExampleControllerWithMiddleware) SayHello(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, "Hello")
	panic("Hello")
}

func (e *ExampleControllerWithMiddleware) SayHi(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, "Hi")
}

func PanicHandlingMiddleware(w http.ResponseWriter, r *http.Request, next http.Handler) {
	slog.Log(r.Context(), slog.LevelInfo, "(PanicHandlingMiddleware) handling panic middleware applied")
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
		slog.Warn("(PanicHandlingMiddleware) panic recovered", slog.String("error", errstring))
	}()
	next.ServeHTTP(w, r)
	slog.Info("(PanicHandlingMiddleware) end")
}

func LoggingMiddlewareFunc(w http.ResponseWriter, r *http.Request, next http.Handler) {
	startTime := time.Now()
	slog.Log(r.Context(), slog.LevelInfo, "(LoggingMiddleware) handling request for: "+r.RequestURI)
	next.ServeHTTP(w, r)
	d := time.Since(startTime)
	slog.Log(r.Context(), slog.LevelInfo, "(LoggingMiddleware) end request for: "+r.RequestURI+" elapsed: "+d.String())
}

func main() {
	mux := http.NewServeMux()
	controller.MustRegisterController(mux,
		&ExampleControllerWithMiddleware{},
		controller.WithMiddlewares(
			middleware.MiddlewareFunc(LoggingMiddlewareFunc),
		),
		controller.WithMiddlewares(
			middleware.MiddlewareFunc(PanicHandlingMiddleware),
		),
	)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
