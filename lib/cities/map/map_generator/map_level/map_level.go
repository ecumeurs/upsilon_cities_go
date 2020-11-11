package map_level

//GeneratorLevel tell what is overridable and what's not.
//Means what's set on a level can't be removed by other levels, with some exceptions ...
type GeneratorLevel int

const (
	Ground         GeneratorLevel = 0 // Sea, Mountains
	River          GeneratorLevel = 1 // River rolls from mountains to seas
	Landscape      GeneratorLevel = 2 // this is what we may find elsewhere ( forest, desert )
	Resource       GeneratorLevel = 3 // Ressource assignation: Note this is mostly macro assignation of ressource (like here are minerals, here are plants, with exceptions )
	Structure      GeneratorLevel = 4 // Structures, like Cities, may be set a bit anywhere ... with some exceptions.
	Transportation GeneratorLevel = 5 // Transportation level ( roads, mostly ) will only be applied between cities
)
