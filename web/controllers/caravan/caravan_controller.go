package caravan_controller

import "net/http"

//Index GET /caravan List all caravans for city's owner.
func Index(w http.ResponseWriter, req *http.Request) {

}

//New GET /caravan/new/:city_id allow to initiate caravan.
func New(w http.ResponseWriter, req *http.Request) {

}

//Create POST /caravan details of caravan.
func Create(w http.ResponseWriter, req *http.Request) {

}

//Seek GET /caravan/seek seek cities candidate for provided items.
func Seek(w http.ResponseWriter, req *http.Request) {

}

//Show GET /caravan/:crv_id details of caravan.
func Show(w http.ResponseWriter, req *http.Request) {

}

//Accept POST /caravan/:crv_id/accept accept new contract
func Accept(w http.ResponseWriter, req *http.Request) {

}

//Reject POST /caravan/:crv_id/reject reject contract
func Reject(w http.ResponseWriter, req *http.Request) {

}

//Abort POST /caravan/:crv_id/abort abort caravan
func Abort(w http.ResponseWriter, req *http.Request) {

}

//GetCounter GET /caravan/:crv_id/counter propose counter proposition
func GetCounter(w http.ResponseWriter, req *http.Request) {

}

//PostCounter POST /caravan/:crv_id/counter propose counter proposition
func PostCounter(w http.ResponseWriter, req *http.Request) {

}

//Drop POST /caravan/:crv_id/drop abort and remove from display.
func Drop(w http.ResponseWriter, req *http.Request) {

}
