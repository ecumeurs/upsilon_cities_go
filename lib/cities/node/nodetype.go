package node

type NodeType int

const (
	None     NodeType = 0 // not used for path finding
	Plain    NodeType = 1 // plain nothing ;)
	CityNode NodeType = 2 // well, cities ;)
	Road     NodeType = 3 // pathways
	Sea      NodeType = 4 // unpassable
	Mountain NodeType = 5 // unpassable
	Forest   NodeType = 6
	River    NodeType = 7
	Desert   NodeType = 8
)

func (node NodeType) String() string {
	names := [...]string{
		"None",
		"City",
		"Road",
		"Sea",
		"Mountain",
	}

	if node < None || node > Mountain {
		return "Unknown"
	}

	return names[node]
}

//Short short name of the node for display.
func (node NodeType) Short() string {
	names := [...]string{
		".",
		"C",
		"R",
		"S",
		"M",
	}

	if node < None || node > Mountain {
		return "?"
	}

	return names[node]
}
