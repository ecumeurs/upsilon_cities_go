package node

type NodeType int

const (
	None     NodeType = 0 // not used for path finding
	CityNode NodeType = 1 // well, cities ;)
	Road     NodeType = 2 // pathways
	Sea      NodeType = 3 // unpassable
	Mountain NodeType = 4 // unpassable
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
