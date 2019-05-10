package corporation

type Corporation struct {
	ID       int
	Name     string
	GridID   int
	CitiesID []int

	Compteur int
}

func New() (corporation *Corporation){
	corporation = new(Corporation)
	corporation.Compteur++
	corporation.ID = corporation.Compteur



}
