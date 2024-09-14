# Go HTTP Controller Registration

This project provides functionality to register controllers for http.ServeMux or any other compatible multiplexers.
It simplifies the process of organizing and managing HTTP handlers in your Go web applications.

## Usage

Here’s a simple example of how to register controllers:

```go
package main

import (
	"io"
	"net/http"

	"github.com/LiquidTheDangerous/commonhttp/controller"
)

type ExampleController struct {
}

func (e *ExampleController) Routes() controller.Routes {
	return []controller.RouteDef{controller.Route("GET", "/hello", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello, world")
	})}
}

func main() {
	mux := http.NewServeMux()
	controller.MustRegisterController(mux, &ExampleController{})
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}

```

### Creating a Controller

You can create a controller by implementing the Controller interface.
The handler for a Route can be a function or an object implementing http.Handler.

```go

var handler http.Handler
var f func(w http.ResponseWriter, r *http.Request)
...
func (c *MyController) Routes() controller.Routes {
    return controller.Routes{controller.Route("GET", "/handler", handler), controller.Route("GET", "/function", f)}
}
...
```

## Using gorilla mux

```go
package main
// create function, that registers route to gorilla router
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
	// Provide this function via option `WithHandlerRegistrarFunc`
    controller.MustRegisterController(m, &MyController{}, controller.WithHandlerRegistrarFunc(RegisterGorillaMux))
    if err := http.ListenAndServe(":8080", m); err != nil {
        panic(err)
    }
}

```

## Contributing

Contributions are welcome! 
If you have suggestions for improvements or find bugs, feel free to open an issue or submit a pull request.
