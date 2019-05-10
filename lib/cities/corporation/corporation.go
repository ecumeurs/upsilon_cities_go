package corporation

type Corporation struct {
	ID       int
	Name     string
	GridID   int
	CitiesID []int

	Compteur int
}

//New create a new corporation.
func New() (corporation *Corporation) {
	corporation = new(Corporation)
	corporation.Compteur++
	corporation.ID = corporation.Compteur
	return corporation
}
