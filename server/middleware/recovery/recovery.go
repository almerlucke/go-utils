package recovery

// Recovery is a Negroni middleware that recovers from any panics and writes a 500 if there was one.
import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/almerlucke/go-utils/server/response"
)

// Middleware middleware
type Middleware struct {
	Logger           *log.Logger
	PrintStack       bool
	ErrorHandlerFunc func(interface{})
	StackAll         bool
	StackSize        int
}

// New returns a new instance of recovery middlewar
func New() *Middleware {
	return &Middleware{
		Logger:     log.New(os.Stdout, "[recovery] ", 0),
		PrintStack: true,
		StackAll:   false,
		StackSize:  1024 * 8,
	}
}

func (ware *Middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func() {
		if err := recover(); err != nil {

			stack := make([]byte, ware.StackSize)
			stack = stack[:runtime.Stack(stack, ware.StackAll)]

			f := "PANIC: %s\n%s"
			ware.Logger.Printf(f, err, stack)

			if ware.PrintStack {
				response.InternalServerError(rw, fmt.Sprintf(f, err, stack))
			} else {
				response.InternalServerError(rw, "")
			}

			if ware.ErrorHandlerFunc != nil {
				func() {
					defer func() {
						if innerErr := recover(); innerErr != nil {
							ware.Logger.Printf("provided ErrorHandlerFunc panic'd: %s, trace:\n%s", innerErr, debug.Stack())
							ware.Logger.Printf("%s\n", debug.Stack())
						}
					}()
					ware.ErrorHandlerFunc(err)
				}()
			}
		}
	}()

	next(rw, r)
}
