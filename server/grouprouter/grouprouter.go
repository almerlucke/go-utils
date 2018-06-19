// Package grouprouter uses httprouter Lookup to create a grouped router where each
// group can use it's own middleware. So we do not need to check on prefix path or any
// such thing. If the lookup for a group returns a handle we call the middleware stack of the
// matching httprouter. You can add a self created group or create a new group with router and
// middleware added. We use httprouter for routing and negroni for middleware
package grouprouter

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

// Group a router and middleware together
type Group struct {
	Router     *httprouter.Router
	Middleware *negroni.Negroni
}

// NewGroup creates a new group
func NewGroup() *Group {
	g := &Group{
		Middleware: negroni.New(),
		Router:     httprouter.New(),
	}

	return g
}

// Prepare group for final use by adding router as last handler
func (g *Group) Prepare() {
	g.Middleware.UseHandler(g.Router)
}

// GroupRouter is a wrapper around one or more middleware and httprouter groups
type GroupRouter struct {
	Groups   []*Group
	Fallback http.Handler
}

// NewGroupRouter creates a new router
func NewGroupRouter(fallback http.Handler) *GroupRouter {
	return &GroupRouter{
		Groups:   []*Group{},
		Fallback: fallback,
	}
}

// AddNewGroup adds a new group
func (r *GroupRouter) AddNewGroup() *Group {
	g := NewGroup()
	r.Groups = append(r.Groups, g)

	return g
}

// AddGroup adds an existing group
func (r *GroupRouter) AddGroup(g *Group) {
	r.Groups = append(r.Groups, g)
}

// ServeHTTP serve the http
func (r *GroupRouter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	method := req.Method
	path := req.URL.Path

	// For each group check if the httprouter handles the path
	// if the router lookup succeeds we call the ServeHTTP of the
	// groups middleware
	for _, g := range r.Groups {
		h, _, _ := g.Router.Lookup(method, path)
		if h != nil {
			g.Middleware.ServeHTTP(rw, req)
			return
		}
	}

	// Call router fallback handler if we didn't find the method/path combination in
	// one of the groups
	if r.Fallback != nil {
		r.Fallback.ServeHTTP(rw, req)
	}
}

/*
EXAMPLE USAGE:

func sharedCors() *cors.Cors {
	return cors.New(cors.Options{
		AllowedMethods: []string{"POST", "GET", "OPTIONS", "DELETE", "PUT"},
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		ExposedHeaders: []string{"Neurokeys-XToken-Validate", "Neurokeys-XToken"},
	})
}

func sharedLocalization() *localization.Middleware {
	return localization.New(language.NewMatcher([]language.Tag{
		language.English, // The first language is used as fallback.
	}))
}

func publicMiddleware(router *httprouter.Router) *negroni.Negroni {
	n := negroni.New()

	n.Use(sharedLocalization())
	n.Use(negroni.NewLogger())
	n.Use(recovery.New())
	n.Use(gzip.Gzip(gzip.DefaultCompression))
	n.Use(sharedCors())
	n.UseHandler(router)

	return n
}

func privateMiddleware(router *httprouter.Router) *negroni.Negroni {
	n := negroni.New()

	n.Use(sharedLocalization())
	n.Use(authtoken.New(&auth.TokenDataFactory{}, "test"))
	n.Use(negroni.NewLogger())
	n.Use(recovery.New())
	n.Use(gzip.Gzip(gzip.DefaultCompression))
	n.Use(sharedCors())
	n.UseHandler(router)

	return n
}



func main() {
	// Seed random
	rand.Seed(time.Now().UTC().UnixNano())

	publicRouter := httprouter.New()
	publicRouter.POST("/api/v1/public/login", func(rw http.ResponseWriter, r *http.Request, pm httprouter.Params) {
		loc, ok := localization.GetLocalization(r.Context())
		if ok {
			log.Printf("public loc %v", loc)
		}
	})

	r1 := publicMiddleware(publicRouter)

	privateRouter := httprouter.New()
	privateRouter.POST("/api/v1/private/check", func(rw http.ResponseWriter, r *http.Request, pm httprouter.Params) {
		loc, ok := localization.GetLocalization(r.Context())
		if ok {
			log.Printf("privateloc %v", loc)
		}
	})

	r2 := privateMiddleware(privateRouter)

	r := NewGroupRouter(nil)

	r.AddGroup(&Group{
		Router:     publicRouter,
		Middleware: r1,
	})

	r.AddGroup(&Group{
		Router:     privateRouter,
		Middleware: r2,
	})

	// Start server
	server := &http.Server{
		Addr:           fmt.Sprintf(":%v", os.Getenv("GLOTTY_PORT")),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
*/
