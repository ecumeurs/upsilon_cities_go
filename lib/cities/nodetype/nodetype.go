package nodetype

import (
	"bytes"
	"encoding/json"
)

type NodeType int

const (
	None         NodeType = 0 // not used for path finding
	Plain        NodeType = 1 // plain nothing ;)
	CityNode     NodeType = 2 // well, cities ;)
	Road         NodeType = 3 // pathways
	Sea          NodeType = 4 // unpassable
	Mountain     NodeType = 5 // unpassable
	Forest       NodeType = 6
	River        NodeType = 7
	Desert       NodeType = 8
	Accessible   NodeType = 9
	Inaccessible NodeType = 10
)

var toEnum = map[string]NodeType{
	"None":         None,
	"Plain":        Plain,
	"City":         CityNode,
	"Road":         Road,
	"Sea":          Sea,
	"Mountain":     Mountain,
	"Forest":       Forest,
	"River":        River,
	"Desert":       Desert,
	"Accessible":   Accessible,
	"Inaccessible": Inaccessible,
}

var names = [...]string{
	"None",
	"Plain",
	"City",
	"Road",
	"Sea",
	"Mountain",
	"Forest",
	"River",
	"Desert",
	"Accessible",
	"Inaccessible",
}

var shortnames = [...]string{
	".",
	"P",
	"C",
	"R",
	"S",
	"M",
	"F",
	"R",
	"D",
	".",
	"X",
}

func (node NodeType) String() string {

	if node < None || node > Inaccessible {
		return "Unknown"
	}

	return names[node]
}

//Short short name of the node for display.
func (node NodeType) Short() string {

	if node < None || node > Inaccessible {
		return "?"
	}

	return shortnames[node]
}

//FromString convert value to node type
func FromString(n string) NodeType {

	return None
}

// MarshalJSON marshals the enum as a quoted json string
func (node NodeType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(node.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (node *NodeType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Created' in this case.
	*node = toEnum[j]
	return nil
}
