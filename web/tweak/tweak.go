//Package tweak circular dependency counter ... -_- :'( :'( :'( :'(
package tweak

import "github.com/gorilla/mux"

var defaultRouter *mux.Router

//GetRouter router.
func GetRouter() *mux.Router {
	return defaultRouter
}

//SetRouter router.
func SetRouter(m *mux.Router) {
	defaultRouter = m
}
