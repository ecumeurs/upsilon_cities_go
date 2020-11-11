package resource

import (
	"fmt"
	"strings"
	"upsilon_cities_go/lib/cities/nodetype"
)

//Resource describe a ressource type that may be harvested from a cell.
//It limits available producers
type Resource struct {
	ID          int `json:"-"`
	Type        string
	Name        string
	Constraints []Constraint
	Rarity      int  // used to build roll table - 1 being quite rare, 10 being common
	Exclusive   bool // if set to true, will mark this resource as being excluisve, only one of this type may be present on the roll table. Prevent proliferation. In case of multiple items only the least rare will be used.

}

//Constraint describe requirements for a resource to be available on a cell
type Constraint struct {
	NodeType  nodetype.NodeType // node type
	Depth     int               // how many layers of this node type must be present to allow this constraint to be lifted.
	Proximity int               // distance allowed to find required node 0 means it must be this cell
}

func (r Resource) String() string {
	exc := ""
	if r.Exclusive {
		exc = "X"
	}
	return strings.TrimSpace(fmt.Sprintf("R %v(%v - %d) C %v %v", r.Name, r.Type, r.Rarity, r.ShortConstraints(), exc))
}

//ShortConstraints short string of constraints
func (r Resource) ShortConstraints() string {
	var arr []string
	for _, v := range r.Constraints {
		arr = append(arr, fmt.Sprintf("%v(%d;%d)", v.NodeType.Short(), v.Depth, v.Proximity))
	}

	return strings.Join(arr, ",")
}

//Clean removes constraints for a nice resource to be stored in database.
func (r *Resource) Clean() {
	r.Constraints = r.Constraints[:]
}
