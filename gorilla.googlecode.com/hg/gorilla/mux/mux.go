// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"bytes"
	"fmt"
	"http"
	"os"
	"path"
	"regexp"
	"strings"
	"url"
	"gorilla.googlecode.com/hg/gorilla/context"
)

// All error descriptions.
const (
	errRouteName string = "Duplicated route name: %q."
	errVarName   string = "Duplicated route variable name: %q."
	// Template parsing.
	errUnbalancedBraces string = "Unbalanced curly braces in route template: %q."
	errBadTemplatePart  string = "Missing name or pattern in route template: %q."
	// URL building.
	errMissingRouteVar string = "Missing route variable: %q."
	errBadRouteVar     string = "Route variable doesn't match: got %q, expected %q."
	errMissingHost     string = "Route doesn't have a host."
	errMissingPath     string = "Route doesn't have a path."
	// Empty parameter errors.
	errEmptyHost       string = "Host() requires a non-zero string, got %q."
	errEmptyPath       string = "Path() requires a non-zero string that starts with a slash, got %q."
	errEmptyPathPrefix string = "PathPrefix() requires a non-zero string that starts with a slash, got %q."
	// Variadic errors.
	errEmptyHeaders string = "Headers() requires at least a pair of parameters."
	errEmptyMethods string = "Methods() requires at least one parameter."
	errEmptyQueries string = "Queries() requires at least a pair of parameters."
	errEmptySchemes string = "Schemes() requires at least one parameter."
	errOddHeaders   string = "Headers() requires an even number of parameters, got %v."
	errOddQueries   string = "Queries() requires an even number of parameters, got %v."
	errOddURLPairs  string = "URL() requires an even number of parameters, got %v."
)

// ----------------------------------------------------------------------------
// Context
// ----------------------------------------------------------------------------

// Vars stores the variables extracted from a URL.
type RouteVars map[string]string

// ctx is the request context namespace for this package.
//
// It stores route variables for each request.
var ctx = new(context.Namespace)

// Vars returns the route variables for the matched route in a given request.
func Vars(request *http.Request) RouteVars {
	rv := ctx.Get(request)
	if rv != nil {
		return rv.(RouteVars)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Router
// ----------------------------------------------------------------------------

// Router registers routes to be matched and dispatches a handler.
//
// It implements the http.Handler interface, so it can be registered to serve
// requests. For example, to send all incoming requests to the default router:
//
//     func main() {
//         http.Handle("/", mux.DefaultRouter)
//     }
//
// Or, for Google App Engine, register it in a init() function:
//
//     func init() {
//         http.Handle("/", mux.DefaultRouter)
//     }
//
// The DefaultRouter is a Router instance ready to register URLs and handlers.
// If needed, new instances can be created and registered.
type Router struct {
	// Routes to be matched, in order.
	Routes []*Route
	// Routes by name, for URL building.
	NamedRoutes map[string]*Route
	// Reference to the root router, where named routes are stored.
	rootRouter *Router
	// Configurable Handler to be used when no route matches.
	NotFoundHandler http.Handler
	// See Route.redirectSlash. This defines the default flag for new routes.
	redirectSlash bool
}

// root returns the root router, where named routes are stored.
func (r *Router) root() *Router {
	if r.rootRouter == nil {
		return r
	}
	return r.rootRouter
}

// Match matches registered routes against the request.
func (r *Router) Match(request *http.Request) (match *RouteMatch, ok bool) {
	for _, route := range r.Routes {
		if match, ok = route.Match(request); ok {
			return
		}
	}
	return
}

// ServeHTTP dispatches the handler registered in the matched route.
//
// When there is a match, the route variables can be retrieved calling
// mux.Vars(request).
func (r *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// Clean path to canonical form and redirect.
	// (this comes from the http package)
	if p := cleanPath(request.URL.Path); p != request.URL.Path {
		writer.Header().Set("Location", p)
		writer.WriteHeader(http.StatusMovedPermanently)
		return
	}
	var handler http.Handler
	if match, ok := r.Match(request); ok {
		handler = match.Handler
	}
	if handler == nil {
		if r.NotFoundHandler == nil {
			r.NotFoundHandler = http.NotFoundHandler()
		}
		handler = r.NotFoundHandler
	}
	defer context.DefaultContext.Clear(request)
	handler.ServeHTTP(writer, request)
}

// AddRoute registers a route in the router.
func (r *Router) AddRoute(route *Route) *Router {
	if r.Routes == nil {
		r.Routes = make([]*Route, 0)
	}
	route.router = r
	r.Routes = append(r.Routes, route)
	return r
}

// RedirectSlash defines the default RedirectSlash behavior for new routes.
//
// See Route.RedirectSlash.
func (r *Router) RedirectSlash(value bool) *Router {
	r.redirectSlash = value
	return r
}

// Convenience route factories ------------------------------------------------

// NewRoute creates an empty route and registers it in the router.
func (r *Router) NewRoute() *Route {
	route := newRoute()
	route.redirectSlash = r.redirectSlash
	r.AddRoute(route)
	return route
}

// Handler registers a new route and sets a handler.
//
// See also: Route.Handler().
func (r *Router) Handler(handler http.Handler) *Route {
	return r.NewRoute().Handler(handler)
}

// HandlerFunc registers a new route and sets a handler function.
//
// See also: Route.HandlerFunc().
func (r *Router) HandlerFunc(handler func(http.ResponseWriter,
*http.Request)) *Route {
	return r.NewRoute().HandlerFunc(handler)
}

// Handle registers a new route and sets a path and handler.
//
// See also: Route.Handle().
func (r *Router) Handle(path string, handler http.Handler) *Route {
	return r.NewRoute().Handle(path, handler)
}

// HandleFunc registers a new route and sets a path and handler function.
//
// See also: Route.HandleFunc().
func (r *Router) HandleFunc(path string, handler func(http.ResponseWriter,
*http.Request)) *Route {
	return r.NewRoute().HandleFunc(path, handler)
}

// Name registers a new route and sets the route name.
//
// See also: Route.Name().
func (r *Router) Name(name string) *Route {
	return r.NewRoute().Name(name)
}

// Convenience route matcher factories ----------------------------------------

// Headers registers a new route and sets a headers matcher.
//
// See also: Route.Headers().
func (r *Router) Headers(pairs ...string) *Route {
	return r.NewRoute().Headers(pairs...)
}

// Host registers a new route and sets a host matcher.
//
// See also: Route.Host().
func (r *Router) Host(template string) *Route {
	return r.NewRoute().Host(template)
}

// Matcher registers a new route and sets a custom matcher function.
//
// See also: Route.Matcher().
func (r *Router) Matcher(matcherFunc MatcherFunc) *Route {
	return r.NewRoute().Matcher(matcherFunc)
}

// Methods registers a new route and sets a methods matcher.
//
// See also: Route.Methods().
func (r *Router) Methods(methods ...string) *Route {
	return r.NewRoute().Methods(methods...)
}

// Path registers a new route and sets a path matcher.
//
// See also: Route.Path().
func (r *Router) Path(template string) *Route {
	return r.NewRoute().Path(template)
}

// PathPrefix registers a new route and sets a path prefix matcher.
//
// See also: Route.PathPrefix().
func (r *Router) PathPrefix(template string) *Route {
	return r.NewRoute().PathPrefix(template)
}

// Queries registers a new route and sets a queries matcher.
//
// See also: Route.Queries().
func (r *Router) Queries(pairs ...string) *Route {
	return r.NewRoute().Queries(pairs...)
}

// Schemes registers a new route and sets a schemes matcher.
//
// See also: Route.Schemes().
func (r *Router) Schemes(schemes ...string) *Route {
	return r.NewRoute().Schemes(schemes...)
}

// ----------------------------------------------------------------------------
// Default Router
// ----------------------------------------------------------------------------
// There's no new logic here, only functions that mirror Router methods to
// change the default router instance.

// DefaultRouter is a default Router instance for convenience.
var DefaultRouter = &Router{
	NamedRoutes:   make(map[string]*Route),
	redirectSlash: true,
}

// NamedRoutes is the DefaultRouter's NamedRoutes field.
var NamedRoutes = DefaultRouter.NamedRoutes

// AddRoute registers a route in the default router.
func AddRoute(route *Route) *Router {
	return DefaultRouter.AddRoute(route)
}

// Route factories ------------------------------------------------------------

// NewRoute creates an empty route and registers it in the router.
func NewRoute() *Route {
	return DefaultRouter.NewRoute()
}

// Handler registers a new route and sets a handler.
//
// See also: Route.Handler().
func Handler(handler http.Handler) *Route {
	return DefaultRouter.NewRoute().Handler(handler)
}

// HandlerFunc registers a new route and sets a handler function.
//
// See also: Route.HandlerFunc().
func HandlerFunc(handler func(http.ResponseWriter, *http.Request)) *Route {
	return DefaultRouter.NewRoute().HandlerFunc(handler)
}

// Handle registers a new route and sets a path and handler.
//
// See also: Route.Handle().
func Handle(path string, handler http.Handler) *Route {
	return DefaultRouter.NewRoute().Handle(path, handler)
}

// HandleFunc registers a new route and sets a path and handler function.
//
// See also: Route.HandleFunc().
func HandleFunc(path string, handler func(http.ResponseWriter,
*http.Request)) *Route {
	return DefaultRouter.NewRoute().HandleFunc(path, handler)
}

// Name registers a new route and sets the route name.
//
// See also: Route.Name().
func Name(name string) *Route {
	return DefaultRouter.NewRoute().Name(name)
}

// Route matcher factories ----------------------------------------------------

// Headers registers a new route and sets a headers matcher.
//
// See also: Route.Headers().
func Headers(pairs ...string) *Route {
	return DefaultRouter.NewRoute().Headers(pairs...)
}

// Host registers a new route and sets a host matcher.
//
// See also: Route.Host().
func Host(template string) *Route {
	return DefaultRouter.NewRoute().Host(template)
}

// Matcher registers a new route and sets a custom matcher function.
//
// See also: Route.Matcher().
func Matcher(matcherFunc MatcherFunc) *Route {
	return DefaultRouter.NewRoute().Matcher(matcherFunc)
}

// Methods registers a new route and sets a methods matcher.
//
// See also: Route.Methods().
func Methods(methods ...string) *Route {
	return DefaultRouter.NewRoute().Methods(methods...)
}

// Path registers a new route and sets a path matcher.
//
// See also: Route.Path().
func Path(template string) *Route {
	return DefaultRouter.NewRoute().Path(template)
}

// PathPrefix registers a new route and sets a path prefix matcher.
//
// See also: Route.PathPrefix().
func PathPrefix(template string) *Route {
	return DefaultRouter.NewRoute().PathPrefix(template)
}

// Queries registers a new route and sets a queries matcher.
//
// See also: Route.Queries().
func Queries(pairs ...string) *Route {
	return DefaultRouter.NewRoute().Queries(pairs...)
}

// Schemes registers a new route and sets a schemes matcher.
//
// See also: Route.Schemes().
func Schemes(schemes ...string) *Route {
	return DefaultRouter.NewRoute().Schemes(schemes...)
}

// ----------------------------------------------------------------------------
// Route
// ----------------------------------------------------------------------------

// Route stores information to match a request.
type Route struct {
	// Reference to the router.
	router *Router
	// Request handler for this route.
	handler http.Handler
	// List of matchers.
	matchers []*routeMatcher
	// Special case matcher: parsed template for host matching.
	hostTemplate *parsedTemplate
	// Special case matcher: parsed template for path matching.
	pathTemplate *parsedTemplate
	// Redirect access from paths not ending with slash to the slash'ed path
	// if the Route paths ends with a slash, and vice-versa.
	// If pattern is /path/, insert permanent redirect for /path.
	redirectSlash bool
}

// newRoute returns a new Route instance.
func newRoute() *Route {
	return &Route{
		matchers: make([]*routeMatcher, 0),
	}
}

// Clone clones a route.
func (r *Route) Clone() *Route {
	// Fields are private and not changed once set, so we can reuse matchers
	// and parsed templates. Must make a copy of the matchers array, though.
	matchers := make([]*routeMatcher, len(r.matchers))
	copy(matchers, r.matchers)
	return &Route{
		router:       r.router,
		handler:      r.handler,
		matchers:     matchers,
		hostTemplate: r.hostTemplate,
		pathTemplate: r.pathTemplate,
		redirectSlash:  r.redirectSlash,
	}
}

// Match matches this route against the request.
//
// It sets variables from the matched route in the context, if any.
func (r *Route) Match(req *http.Request) (*RouteMatch, bool) {
	var hostMatches, pathMatches []string
	if r.hostTemplate != nil {
		hostMatches = r.hostTemplate.Regexp.FindStringSubmatch(req.URL.Host)
		if hostMatches == nil {
			return nil, false
		}
	}
	var redirectURL string
	if r.pathTemplate != nil {
		// TODO Match the path unescaped?
		/*
			if path, ok := url.URLUnescape(r.URL.Path); ok {
				// URLUnescape converts '+' into ' ' (space). Revert this.
				path = strings.Replace(path, " ", "+", -1)
			} else {
				path = r.URL.Path
			}
		*/
		pathMatches = r.pathTemplate.Regexp.FindStringSubmatch(req.URL.Path)
		if pathMatches == nil {
			return nil, false
		} else if r.redirectSlash {
			// Check if we should redirect.
			p1 := strings.HasSuffix(req.URL.Path, "/")
			p2 := strings.HasSuffix(r.pathTemplate.Template, "/")
			if p1 != p2 {
				ru, _ := url.Parse(req.URL.String())
				if p1 {
					ru.Path = ru.Path[:len(ru.Path)-1]
				} else {
					ru.Path = ru.Path + "/"
				}
				redirectURL = ru.String()
			}
		}
	}
	var match *RouteMatch
	if r.matchers != nil {
		for _, matcher := range r.matchers {
			if rv, ok := (*matcher).Match(req); !ok {
				return nil, false
			} else if rv != nil {
				match = rv
				break
			}
		}
	}
	// We have a match.
	vars := make(RouteVars)
	if hostMatches != nil {
		for k, v := range r.hostTemplate.VarsN {
			vars[v] = hostMatches[k+1]
		}
	}
	if pathMatches != nil {
		for k, v := range r.pathTemplate.VarsN {
			vars[v] = pathMatches[k+1]
		}
	}
	ctx.Set(req, vars)
	if match == nil {
		match = &RouteMatch{Route: r, Handler: r.handler}
	}
	if redirectURL != "" {
		match.Handler = http.RedirectHandler(redirectURL, 301)
	}
	return match, true
}

// Subrouting -----------------------------------------------------------------

// NewRouter creates a new router and adds it as a matcher for this route.
//
// This is used for subrouting: it will test the inner routes if other
// matchers matched. For example:
//
//     subrouter := mux.Host("www.domain.com").NewRouter()
//     subrouter.HandleFunc("/products/", ProductsHandler)
//     subrouter.HandleFunc("/products/{key}", ProductHandler)
//     subrouter.HandleFunc("/articles/{category}/{id:[0-9]+}"),
//                          ArticleHandler)
//
// In this example, the routes registered in the subrouter will only be tested
// if the host matches.
func (r *Route) NewRouter() *Router {
	router := &Router{
		Routes:     make([]*Route, 0),
		rootRouter: r.router.root(),
	}
	r.addMatcher(router)
	return router
}

// URL building ---------------------------------------------------------------

// URL builds a URL for the route.
//
// It accepts a sequence of key/value pairs for the route variables. For
// example, given this route:
//
//     mux.HandleFunc("/articles/{category}/{id:[0-9]+}", ArticleHandler).
//         Name("article")
//
// ...a URL for it can be built using:
//
//     url := mux.NamedRoutes["article"].URL("category", "technology",
//                                           "id", "42")
//
// ...which will return an url.URL with the following path:
//
//     "/articles/technology/42"
//
// This also works for host variables:
//
//     mux.Host("{subdomain}.domain.com").
//              HandleFunc("/articles/{category}/{id:[0-9]+}", ArticleHandler).
//              Name("article")
//
//     // url.String() will be "http://news.domain.com/articles/technology/42"
//     url := mux.NamedRoutes["article"].URL("subdomain", "news",
//                                           "category", "technology",
//                                           "id", "42")
//
// All variable names defined in the route are required, and their values must
// conform to the corresponding patterns, if any.
//
// In case of bad arguments it will return nil.
func (r *Route) URL(pairs ...string) (rv *url.URL) {
	rv, _ = r.URLDebug(pairs...)
	return
}

// URLDebug is a debug version of URL: it also returns an error in case of
// failure.
func (r *Route) URLDebug(pairs ...string) (rv *url.URL, err os.Error) {
	var scheme, host, path string
	values := stringMapFromPairs(errOddURLPairs, pairs...)
	if r.hostTemplate != nil {
		// Set a default scheme.
		scheme = "http"
		if host, err = reverseRoute(r.hostTemplate, values); err != nil {
			return
		}
	}
	if r.pathTemplate != nil {
		if path, err = reverseRoute(r.pathTemplate, values); err != nil {
			return
		}
	}
	rv = &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}
	return
}

// URLHost builds the host part of the URL for a route.
//
// The route must have a host defined.
//
// In case of bad arguments or missing host it will return nil.
func (r *Route) URLHost(pairs ...string) (rv *url.URL) {
	rv, _ = r.URLHostDebug(pairs...)
	return
}

// URLHostDebug is a debug version of URLHost: it also returns an error in
// case of failure.
func (r *Route) URLHostDebug(pairs ...string) (rv *url.URL, err os.Error) {
	if r.hostTemplate == nil {
		err = muxError(errMissingHost)
		return
	}
	var host string
	values := stringMapFromPairs(errOddURLPairs, pairs...)
	if host, err = reverseRoute(r.hostTemplate, values); err != nil {
		return
	} else {
		rv = &url.URL{
			Scheme: "http",
			Host:   host,
		}
	}
	return
}

// URLPath builds the path part of the URL for a route.
//
// The route must have a path defined.
//
// In case of bad arguments or missing path it will return nil.
func (r *Route) URLPath(pairs ...string) (rv *url.URL) {
	rv, _ = r.URLPathDebug(pairs...)
	return
}

// URLPathDebug is a debug version of URLPath: it also returns an error in
// case of failure.
func (r *Route) URLPathDebug(pairs ...string) (rv *url.URL, err os.Error) {
	if r.pathTemplate == nil {
		err = muxError(errMissingPath)
		return
	}
	var path string
	values := stringMapFromPairs(errOddURLPairs, pairs...)
	if path, err = reverseRoute(r.pathTemplate, values); err != nil {
		return
	} else {
		rv = &url.URL{
			Path: path,
		}
	}
	return
}

// reverseRoute builds a URL part based on the route's parsed template.
func reverseRoute(tpl *parsedTemplate, values map[string]string) (rv string, err os.Error) {
	var value string
	var ok bool
	urlValues := make([]interface{}, len(tpl.VarsN))
	for k, v := range tpl.VarsN {
		if value, ok = values[v]; !ok {
			err = muxError(errMissingRouteVar, v)
			return
		}
		urlValues[k] = value
	}
	rv = fmt.Sprintf(tpl.Reverse, urlValues...)
	if !tpl.Regexp.MatchString(rv) {
		// The URL is checked against the full regexp, instead of checking
		// individual variables. This is faster but to provide a good error
		// message, we check individual regexps if the URL doesn't match.
		for k, v := range tpl.VarsN {
			if !tpl.VarsR[k].MatchString(values[v]) {
				err = muxError(errBadRouteVar, values[v],
					tpl.VarsR[k].String())
				return
			}
		}
	}
	return
}

// Route predicates -----------------------------------------------------------

// Handler sets a handler for the route.
func (r *Route) Handler(handler http.Handler) *Route {
	r.handler = handler
	return r
}

// HandlerFunc sets a handler function for the route.
func (r *Route) HandlerFunc(handler func(http.ResponseWriter,
*http.Request)) *Route {
	return r.Handler(http.HandlerFunc(handler))
}

// Handle sets a path and handler for the route.
func (r *Route) Handle(path string, handler http.Handler) *Route {
	return r.Path(path).Handler(handler)
}

// HandleFunc sets a path and handler function for the route.
func (r *Route) HandleFunc(path string, handler func(http.ResponseWriter,
*http.Request)) *Route {
	return r.Path(path).Handler(http.HandlerFunc(handler))
}

// Name sets the route name, used to build URLs.
//
// A name must be unique for a router. If the name was registered already
// it will be overwritten.
func (r *Route) Name(name string) *Route {
	router := r.router.root()
	if router.NamedRoutes == nil {
		router.NamedRoutes = make(map[string]*Route)
	}
	router.NamedRoutes[name] = r
	return r
}

// RedirectSlash defines the redirectSlash behavior for this route.
//
// When true, if the route path is /path/, accessing /path will redirect to
// /path/, and vice versa.
func (r *Route) RedirectSlash(value bool) *Route {
	r.redirectSlash = value
	return r
}

// Route matchers -------------------------------------------------------------

// addMatcher adds a matcher to the array of route matchers.
func (r *Route) addMatcher(m routeMatcher) *Route {
	r.matchers = append(r.matchers, &m)
	return r
}

// Headers adds a matcher to match the request against header values.
//
// It accepts a sequence of key/value pairs to be matched. For example:
//
//     mux.Headers("Content-Type", "application/json",
//                 "X-Requested-With", "XMLHttpRequest")
//
// The above route will only match if both request header values match.
//
// It the value is an empty string, it will match any value if the key is set.
func (r *Route) Headers(pairs ...string) *Route {
	headers := stringMapFromPairs(errOddHeaders, pairs...)
	if len(headers) == 0 {
		panic(errEmptyHeaders)
	}
	return r.addMatcher(&headerMatcher{headers: headers})
}

// Host adds a matcher to match the request against the URL host.
//
// It accepts a template with zero or more URL variables enclosed by {}.
// Variables can define an optional regexp pattern to me matched:
//
// - {name} matches anything until the next dot.
//
// - {name:pattern} matches the given regexp pattern.
//
// For example:
//
//     mux.Host("www.domain.com")
//     mux.Host("{subdomain}.domain.com")
//     mux.Host("{subdomain:[a-z]+}.domain.com")
//
// Variable names must be unique in a given route. They can be retrieved
// calling mux.Vars(request).
func (r *Route) Host(template string) *Route {
	if template == "" {
		panic(fmt.Sprintf(errEmptyHost, template))
	}

	tpl := &parsedTemplate{Template: template}
	err := parseTemplate(tpl, "[^.]+", false, false,
		variableNames(r.pathTemplate))
	if err != nil {
		panic(err)
	}
	r.hostTemplate = tpl
	return r
}

// Matcher adds a matcher to match the request using a custom function.
func (r *Route) Matcher(matcherFunc MatcherFunc) *Route {
	return r.addMatcher(&customMatcher{matcherFunc: matcherFunc})
}

// Methods adds a matcher to match the request against HTTP methods.
//
// It accepts a sequence of one or more methods to be matched, e.g.:
// "GET", "POST", "PUT".
func (r *Route) Methods(methods ...string) *Route {
	if len(methods) == 0 {
		panic(errEmptyMethods)
	}
	for k, v := range methods {
		methods[k] = strings.ToUpper(v)
	}
	return r.addMatcher(&methodMatcher{methods: methods})
}

// Path adds a matcher to match the request against the URL path.
//
// It accepts a template with zero or more URL variables enclosed by {}.
// Variables can define an optional regexp pattern to me matched:
//
// - {name} matches anything until the next slash.
//
// - {name:pattern} matches the given regexp pattern.
//
// For example:
//
//     mux.Path("/products/").Handler(ProductsHandler)
//     mux.Path("/products/{key}").Handler(ProductsHandler)
//     mux.Path("/articles/{category}/{id:[0-9]+}").
//             Handler(ArticleHandler)
//
// Variable names must be unique in a given route. They can be retrieved
// calling mux.Vars(request).
func (r *Route) Path(template string) *Route {
	if template == "" || template[0] != '/' {
		panic(fmt.Sprintf(errEmptyPath, template))
	}
	tpl := &parsedTemplate{Template: template}
	err := parseTemplate(tpl, "[^/]+", false, r.redirectSlash,
		variableNames(r.hostTemplate))
	if err != nil {
		panic(err)
	}
	r.pathTemplate = tpl
	return r
}

// PathPrefix adds a matcher to match the request against a URL path prefix.
func (r *Route) PathPrefix(template string) *Route {
	if template == "" || template[0] != '/' {
		panic(fmt.Sprintf(errEmptyPathPrefix, template))
	}
	tpl := &parsedTemplate{Template: template}
	err := parseTemplate(tpl, "[^/]+", true, false,
		variableNames(r.hostTemplate))
	if err != nil {
		panic(err)
	}
	r.pathTemplate = tpl
	return r
}

// Queries adds a matcher to match the request against URL query values.
//
// It accepts a sequence of key/value pairs to be matched. For example:
//
//     mux.Queries("foo", "bar",
//                 "baz", "ding")
//
// The above route will only match if the URL contains the defined queries
// values, e.g.: ?foo=bar&baz=ding.
//
// It the value is an empty string, it will match any value if the key is set.
func (r *Route) Queries(pairs ...string) *Route {
	queries := stringMapFromPairs(errOddQueries, pairs...)
	if len(queries) == 0 {
		panic(errEmptyQueries)
	}
	return r.addMatcher(&queryMatcher{queries: queries})
}

// Schemes adds a matcher to match the request against URL schemes.
//
// It accepts a sequence of one or more schemes to be matched, e.g.:
// "http", "https".
func (r *Route) Schemes(schemes ...string) *Route {
	if len(schemes) == 0 {
		panic(errEmptySchemes)
	}
	for k, v := range schemes {
		schemes[k] = strings.ToLower(v)
	}
	return r.addMatcher(&schemeMatcher{schemes: schemes})
}

// ----------------------------------------------------------------------------
// Matchers
// ----------------------------------------------------------------------------

// routeMatch is the returned result when a route matches.
type RouteMatch struct {
	Route   *Route
	Handler http.Handler
}

// MatcherFunc is the type used by custom matchers.
type MatcherFunc func(*http.Request) bool

// routeMatcher is the interface used by the router, route and route matchers.
//
// Only Router and Route actually return a route; it indicates a final match.
// Route matchers return nil and the result from the individual match.
type routeMatcher interface {
	Match(*http.Request) (*RouteMatch, bool)
}

// customMatcher matches the request using a custom matcher function.
type customMatcher struct {
	matcherFunc MatcherFunc
}

func (m *customMatcher) Match(request *http.Request) (*RouteMatch, bool) {
	return nil, m.matcherFunc(request)
}

// headerMatcher matches the request against header values.
type headerMatcher struct {
	headers map[string]string
}

func (m *headerMatcher) Match(request *http.Request) (*RouteMatch, bool) {
	return nil, matchMap(m.headers, request.Header, true)
}

// methodMatcher matches the request against HTTP methods.
type methodMatcher struct {
	methods []string
}

func (m *methodMatcher) Match(request *http.Request) (*RouteMatch, bool) {
	return nil, matchInArray(m.methods, request.Method)
}

// queryMatcher matches the request against URL queries.
type queryMatcher struct {
	queries map[string]string
}

func (m *queryMatcher) Match(request *http.Request) (*RouteMatch, bool) {
	return nil, matchMap(m.queries, request.URL.Query(), false)
}

// schemeMatcher matches the request against URL schemes.
type schemeMatcher struct {
	schemes []string
}

func (m *schemeMatcher) Match(request *http.Request) (*RouteMatch, bool) {
	return nil, matchInArray(m.schemes, request.URL.Scheme)
}

// ----------------------------------------------------------------------------
// Template parsing
// ----------------------------------------------------------------------------

// parsedTemplate stores a regexp and variables info for a route matcher.
type parsedTemplate struct {
	// The unmodified template.
	Template string
	// Expanded regexp.
	Regexp   *regexp.Regexp
	// Reverse template.
	Reverse  string
	// Variable names.
	VarsN    []string
	// Variable regexps (validators).
	VarsR    []*regexp.Regexp
}

// parseTemplate parses a route template, expanding variables into regexps.
//
// It will extract named variables, assemble a regexp to be matched, create
// a "reverse" template to build URLs and compile regexps to validate variable
// values used in URL building.
//
// Previously we accepted only Python-like identifiers for variable
// names ([a-zA-Z_][a-zA-Z0-9_]*), but currently the only restriction is that
// name and pattern can't be empty, and names can't contain a colon.
func parseTemplate(tpl *parsedTemplate, defaultPattern string, prefix bool,
	redirectSlash bool, names *[]string) os.Error {
	// Set a flag for redirectSlash.
	template := tpl.Template
	endSlash := false
	if redirectSlash && strings.HasSuffix(template, "/") {
		template = template[:len(template)-1]
		endSlash = true
	}

	idxs, err := getBraceIndices(template)
	if err != nil {
		return err
	}

	var raw, name, patt string
	var end int
	var parts []string
	pattern := bytes.NewBufferString("^")
	reverse := bytes.NewBufferString("")
	size := len(idxs)
	tpl.VarsN = make([]string, size/2)
	tpl.VarsR = make([]*regexp.Regexp, size/2)
	for i := 0; i < size; i += 2 {
		// 1. Set all values we are interested in.
		raw = template[end:idxs[i]]
		end = idxs[i+1]
		parts = strings.SplitN(template[idxs[i]+1:end-1], ":", 2)
		name = parts[0]
		if len(parts) == 1 {
			patt = defaultPattern
		} else {
			patt = parts[1]
		}
		// Name or pattern can't be empty.
		if name == "" || patt == "" {
			return muxError(errBadTemplatePart, template[idxs[i]:end])
		}
		// Name must be unique for the route.
		if names != nil {
			if matchInArray(*names, name) {
				return muxError(errVarName, name)
			}
			*names = append(*names, name)
		}
		// 2. Build the regexp pattern.
		fmt.Fprintf(pattern, "%s(%s)", regexp.QuoteMeta(raw), patt)
		// 3. Build the reverse template.
		fmt.Fprintf(reverse, "%s%%s", raw)
		// 4. Append variable name and compiled pattern.
		tpl.VarsN[i/2] = name
		if reg, err := regexp.Compile(fmt.Sprintf("^%s$", patt)); err != nil {
			return err
		} else {
			tpl.VarsR[i/2] = reg
		}
	}
	// 5. Add the remaining.
	raw = template[end:]
	pattern.WriteString(regexp.QuoteMeta(raw))
	if redirectSlash {
		pattern.WriteString("[/]?")
	}
	if !prefix {
		pattern.WriteString("$")
	}
	reverse.WriteString(raw)
	if endSlash {
		reverse.WriteString("/")
	}
	// Done!
	reg, err := regexp.Compile(pattern.String())
	if err != nil {
		return err
	}
	tpl.Regexp = reg
	tpl.Reverse = reverse.String()
	return nil
}

// getBraceIndices returns index bounds for route template variables.
//
// It will return an error if there are unbalanced braces.
func getBraceIndices(s string) ([]int, os.Error) {
	var level, idx int
	idxs := make([]int, 0)
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			if level++; level == 1 {
				idx = i
			}
		case '}':
			if level--; level == 0 {
				idxs = append(idxs, idx, i+1)
			} else if level < 0 {
				return nil, muxError(errUnbalancedBraces, s)
			}
		}
	}
	if level != 0 {
		return nil, muxError(errUnbalancedBraces, s)
	}
	return idxs, nil
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// muxError returns a formatted error.
func muxError(msg string, vars ...interface{}) os.Error {
	return os.NewError(fmt.Sprintf(msg, vars...))
}

// cleanPath returns the canonical path for p, eliminating . and .. elements.
//
// Extracted from the http package.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

// stringMapFromPairs converts variadic string parameters to a string map.
func stringMapFromPairs(err string, pairs ...string) map[string]string {
	size := len(pairs)
	if size%2 != 0 {
		panic(fmt.Sprintf(err, pairs))
	}
	m := make(map[string]string, size/2)
	for i := 0; i < size; i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return m
}

// variableNames returns a copy of variable names for route templates.
func variableNames(templates ...*parsedTemplate) *[]string {
	names := make([]string, 0)
	for _, t := range templates {
		if t != nil && len(t.VarsN) != 0 {
			names = append(names, t.VarsN...)
		}
	}
	return &names
}

// matchInArray returns true if the given string value is in the array.
func matchInArray(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

// matchMap returns true if the given key/value pairs exist in a given map.
func matchMap(toCheck map[string]string, toMatch map[string][]string,
canonicalKey bool) bool {
	for k, v := range toCheck {
		// Check if key exists.
		if canonicalKey {
			k = http.CanonicalHeaderKey(k)
		}
		if values, keyExists := toMatch[k]; !keyExists {
			return false
		} else if v != "" {
			// If value was defined as an empty string we only check that the
			// key exists. Otherwise we also check if the value exists.
			valueExists := false
			for _, value := range values {
				if v == value {
					valueExists = true
					break
				}
			}
			if !valueExists {
				return false
			}
		}
	}
	return true
}
