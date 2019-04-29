package cities

type City struct {
	*Node // inherit all node

	Neighbours []City
	Roads      []Pathway
}
