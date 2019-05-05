package generator

import (
	"fmt"
	"math/rand"
)

var bodyList, prefixList, suffixList []string

//Init prepare the whole list for later use ;)
func Init() {
	bodyList = []string{"Arramiguère", "Arrandale", "Arrastèt", "Arraulhè", "Arre", "Arrè", "Arrebentè", "Arrebot", "Arreboulh", "Arrebustalhè", "Arrec", "Arrèc", "Arredau", "Arregalh", "Arrei", "Arrélhe",
		"Arrembès", "Arrémitsa", "Bouès", "Bouésillé", "Bouèsq", "Coudure", "Coue", "Couèou", "Coueyla", "Coufi", "Coufin", "Cougnassa", "Cougnasse", "Cougnét", "Cougnèou", "Cougnot", "Cougourdo", "Couhi", "Couillade", "Coula",
		"Coulancho", "Coulata", "Coulédous", "Coulée", "Dembessè", "Lenguèino", "Lenhous", "Lenire", "Lenn", "Lenz", "Lenza", "Leo", "Lepho", "Lerche", "Lère", "Lès", "Les", "Leschaux", "Lescheraines",
		"Lesco", "Lésine", "Mazière", "Mazuc", "Mazza", "Mé", "Méal", "Mealha", "Neuchli", "Neué", "Neuhatte", "Neule", "Neulo", "Neusance", "Neusière", "Neuva", "Neuyer", "Neuziller", "Nevè", "Nevedenn", "Nevez",
		"Nevezenn", "Nezyié", "Nhesta", "Nié", "Niellu", "Orière", "Orin", "Orle", "Orma", "Ormaie", "Orme", "Reuzeulenn", "Reva", "Revastoulié", "Revelin", "Revers", "Revessen", "Revie", "Reviro", "Revive", "Revola", "Revorenn", "Revorsa",
		"Revou", "Rey", "Rez", "Sestier", "Sestre", "Setérée", "Setier", "Sêtora"}

	prefixList = []string{"Pont-de", "Le-Petit", "Le-Grand", "Petit", "Petite", "La-Petite", "La-Grande",
		"Grand", "Grande", "la-Motte"}

	suffixList = []string{"le-Vieux", "le-Jeune", "la-Jeune", "de-Vals", "du-Serre", "le-Bourg", "le-Chastel",
		"les-Vieilles", "la-Croix", "en-Val", "le-Blanc", "des-Fossés", "les-Fossés",
		"les-Roses", "la-Jolie", "au-Perche", "le-Puy", "Neuf", "le-Haut", "la-Prune",
		"Bellevue", "sur-Sombre", "le-Noir", "les-Courbes",
		"en-l’Air", "la-Joûte", "aux-Miroirs", "aux-Clos", "la-Rouge", "aux-Dames",
		"sur-Trey", "la-Chaussée", "au-Passage"}
}

//CityName Generate a new city name
func CityName() string {

	boolPre := rand.Int31n(100) <= 5
	boolSuf := rand.Int31n(100) <= 5

	name := bodyList[rand.Intn((len(bodyList) - 1))]

	if boolPre {
		prefix := prefixList[rand.Intn((len(prefixList) - 1))]
		name = fmt.Sprintf("%s-%s", prefix, name)
	}

	if boolSuf {
		name = fmt.Sprintf("%s-%s", name, suffixList[rand.Intn(len(suffixList)-1)])
	}

	return name
}
