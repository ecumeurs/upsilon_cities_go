package corporation

type Corporation struct {
	ID       int
	Name     string
	GridID   int
	CitiesID []int

	Compteur int
}

//New create a new corporation.
func New(gridID int) (corporation *Corporation) {
	corporation = new(Corporation)
	corporation.GridID = gridID
	corporation.Compteur++
	corporation.ID = corporation.Compteur
	return corporation
}
