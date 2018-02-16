package router

import (
	"github.com/almerlucke/go-utils/server/auth/jwt"
	"github.com/almerlucke/go-utils/server/handles"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/text/language"
)

// Router - wrapper around httprouter.Router
type Router struct {
	*httprouter.Router
}

// New router
func New() *Router {
	router := &Router{
		Router: httprouter.New(),
	}

	return router
}

/*
   Localization wrapped versions
*/

// GET is a shortcut for router.Handle("GET", path, handle)
func (r *Router) GET(path string, languageMatcher language.Matcher, handle handles.Handle) {
	r.Handle("GET", path, handles.Wrap(languageMatcher, handle))
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (r *Router) HEAD(path string, languageMatcher language.Matcher, handle handles.Handle) {
	r.Handle("HEAD", path, handles.Wrap(languageMatcher, handle))
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (r *Router) OPTIONS(path string, languageMatcher language.Matcher, handle handles.Handle) {
	r.Handle("OPTIONS", path, handles.Wrap(languageMatcher, handle))
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (r *Router) POST(path string, languageMatcher language.Matcher, handle handles.Handle) {
	r.Handle("POST", path, handles.Wrap(languageMatcher, handle))
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (r *Router) PUT(path string, languageMatcher language.Matcher, handle handles.Handle) {
	r.Handle("PUT", path, handles.Wrap(languageMatcher, handle))
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (r *Router) PATCH(path string, languageMatcher language.Matcher, handle handles.Handle) {
	r.Handle("PATCH", path, handles.Wrap(languageMatcher, handle))
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (r *Router) DELETE(path string, languageMatcher language.Matcher, handle handles.Handle) {
	r.Handle("DELETE", path, handles.Wrap(languageMatcher, handle))
}

/*
   JWT auth versions
*/

// JWTAuthGET is a shortcut for authenticated router.Handle("GET", path, handle)
func (r *Router) JWTAuthGET(path string, signingSecret string, languageMatcher language.Matcher, factory jwt.TokenDataFactory, handle handles.JWTAuthHandle) {
	r.Handle("GET", path, handles.JWTAuthWrap(signingSecret, languageMatcher, factory, handle))
}

// JWTAuthHEAD is a shortcut for authenticated router.Handle("HEAD", path, handle)
func (r *Router) JWTAuthHEAD(path string, signingSecret string, languageMatcher language.Matcher, factory jwt.TokenDataFactory, handle handles.JWTAuthHandle) {
	r.Handle("HEAD", path, handles.JWTAuthWrap(signingSecret, languageMatcher, factory, handle))
}

// JWTAuthOPTIONS is a shortcut for authenticated router.Handle("OPTIONS", path, handle)
func (r *Router) JWTAuthOPTIONS(path string, signingSecret string, languageMatcher language.Matcher, factory jwt.TokenDataFactory, handle handles.JWTAuthHandle) {
	r.Handle("OPTIONS", path, handles.JWTAuthWrap(signingSecret, languageMatcher, factory, handle))
}

// JWTAuthPOST is a shortcut for authenticated router.Handle("POST", path, handle)
func (r *Router) JWTAuthPOST(path string, signingSecret string, languageMatcher language.Matcher, factory jwt.TokenDataFactory, handle handles.JWTAuthHandle) {
	r.Handle("POST", path, handles.JWTAuthWrap(signingSecret, languageMatcher, factory, handle))
}

// JWTAuthPUT is a shortcut for authenticated router.Handle("PUT", path, handle)
func (r *Router) JWTAuthPUT(path string, signingSecret string, languageMatcher language.Matcher, factory jwt.TokenDataFactory, handle handles.JWTAuthHandle) {
	r.Handle("PUT", path, handles.JWTAuthWrap(signingSecret, languageMatcher, factory, handle))
}

// JWTAuthPATCH is a shortcut for authenticated router.Handle("PATCH", path, handle)
func (r *Router) JWTAuthPATCH(path string, signingSecret string, languageMatcher language.Matcher, factory jwt.TokenDataFactory, handle handles.JWTAuthHandle) {
	r.Handle("PATCH", path, handles.JWTAuthWrap(signingSecret, languageMatcher, factory, handle))
}

// JWTAuthDELETE is a shortcut for authenticated router.Handle("DELETE", path, handle)
func (r *Router) JWTAuthDELETE(path string, signingSecret string, languageMatcher language.Matcher, factory jwt.TokenDataFactory, handle handles.JWTAuthHandle) {
	r.Handle("DELETE", path, handles.JWTAuthWrap(signingSecret, languageMatcher, factory, handle))
}

/*
   Basic auth versions
*/

// BasicAuthGET is a shortcut for basic authenticated router.Handle("GET", path, handle)
func (r *Router) BasicAuthGET(path string, user string, pwd string, languageMatcher language.Matcher, handle handles.Handle) {
	r.Handle("GET", path, handles.BasicAuthWrap(user, pwd, languageMatcher, handle))
}

// BasicAuthHEAD is a shortcut for basic authenticated router.Handle("HEAD", path, handle)
func (r *Router) BasicAuthHEAD(path string, user string, pwd string, languageMatcher language.Matcher, handle handles.Handle) {
	r.Handle("HEAD", path, handles.BasicAuthWrap(user, pwd, languageMatcher, handle))
}

// BasicAuthOPTIONS is a shortcut for basic authenticated router.Handle("OPTIONS", path, handle)
func (r *Router) BasicAuthOPTIONS(path string, user string, pwd string, languageMatcher language.Matcher, handle handles.Handle) {
	r.Handle("OPTIONS", path, handles.BasicAuthWrap(user, pwd, languageMatcher, handle))
}

// BasicAuthPOST is a shortcut for basic authenticated router.Handle("POST", path, handle)
func (r *Router) BasicAuthPOST(path string, user string, pwd string, languageMatcher language.Matcher, handle handles.Handle) {
	r.Handle("POST", path, handles.BasicAuthWrap(user, pwd, languageMatcher, handle))
}

// BasicAuthPUT is a shortcut for basic authenticated router.Handle("PUT", path, handle)
func (r *Router) BasicAuthPUT(path string, user string, pwd string, languageMatcher language.Matcher, handle handles.Handle) {
	r.Handle("PUT", path, handles.BasicAuthWrap(user, pwd, languageMatcher, handle))
}

// BasicAuthPATCH is a shortcut for basic authenticated router.Handle("PATCH", path, handle)
func (r *Router) BasicAuthPATCH(path string, user string, pwd string, languageMatcher language.Matcher, handle handles.Handle) {
	r.Handle("PATCH", path, handles.BasicAuthWrap(user, pwd, languageMatcher, handle))
}

// BasicAuthDELETE is a shortcut for basic authenticated router.Handle("DELETE", path, handle)
func (r *Router) BasicAuthDELETE(path string, user string, pwd string, languageMatcher language.Matcher, handle handles.Handle) {
	r.Handle("DELETE", path, handles.BasicAuthWrap(user, pwd, languageMatcher, handle))
}
