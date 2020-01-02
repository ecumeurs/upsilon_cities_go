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
		"Plain",
		"City",
		"Road",
		"Sea",
		"Mountain",
		"Forest",
		"River",
		"Desert",
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
		"P",
		"C",
		"R",
		"S",
		"M",
		"F",
		"R",
		"D",
	}

	if node < None || node > Desert {
		return "?"
	}

	return names[node]
}
