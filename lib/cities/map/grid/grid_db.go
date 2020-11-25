package grid

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/db"
)

// DB FUNCTIONS

//Insert grid in database
func (grid *Grid) Insert(dbh *db.Handler) error {
	json, err := grid.dbjsonify()
	if err != nil {
		log.Fatalf("Grid: Failed to jsonify data for database. %s", err)
		return err
	}

	rows, err := dbh.Query("insert into maps(region_name, data) values($1,$2) returning map_id", grid.Name, json)
	if err != nil {
		log.Fatalf("Grid DB: Failed to Insert. %s", err)
		return err
	}
	for rows.Next() {
		rows.Scan(&grid.ID)
	}

	rows.Close()

	log.Printf("Grid: Grid %d Inserted", grid.ID)

	return nil
}

//Update grid in database
func (grid *Grid) Update(dbh *db.Handler) error {
	if grid.ID <= 0 {
		return grid.Insert(dbh)
	}

	json, err := grid.dbjsonify()
	if err != nil {
		log.Fatalf("Grid: Failed to jsonify data for database. %s", err)
		return err
	}

	query, err := dbh.Query(`update maps set
			region_name=$1,
			region_type=$4,
			data=$2,
			updated_at= (now() at time zone 'utc')			
			where map_id=$3;`, grid.Name, json, grid.ID, grid.RegionType)
	if err != nil {
		return fmt.Errorf("Grid DB: Failed to Update Map. %s", err)
	}
	query.Close()
	log.Printf("Grid: Grid %d Updated", grid.ID)
	return nil
}

//Drop grid from database
func (grid *Grid) Drop(dbh *db.Handler) error {
	query, err := dbh.Query("delete from maps where map_id=$1", grid.ID)
	if err != nil {
		return fmt.Errorf("Grid DB: Failed to Drop. %s", err)
	}
	query.Close()
	log.Printf("Grid: Grid %d Deleted", grid.ID)
	grid.ID = 0

	return nil
}

//DropByID grid by ID from database
func DropByID(dbh *db.Handler, id int) error {
	query, err := dbh.Query("delete from maps where map_id=$1", id)
	if err != nil {
		return fmt.Errorf("Grid DB: Failed to DropByID. %s", err)
	}
	query.Close()
	log.Printf("Grid: Grid %d Deleted", id)

	return nil
}

//ByID seek a grid by ID
func ByID(dbh *db.Handler, id int) (grid *Grid, err error) {
	rows, err := dbh.Query("select region_name, region_type, updated_at, data from maps where map_id=$1", id)
	if err != nil {
		return nil, fmt.Errorf("Grid DB: Failed to select map ByID. %s", err)
	}
	for rows.Next() {
		grid := new(Grid)
		grid.Clear()

		var json []byte
		rows.Scan(&grid.Name, &grid.RegionType, &grid.LastUpdate, &json)
		grid.ID = id
		grid.dbunjsonify(json)

		grid.Cities, err = city.ByMap(dbh, id)
		for _, v := range grid.Cities {
			grid.LocationToCity[v.Location.ToInt(grid.Size)] = v
		}

		rows.Close()

		return grid, nil
	}
	return nil, errors.New("Not found")
}

//NameByID seek a Name by ID
func NameByID(dbh *db.Handler, id int) (string, error) {

	rows, err := dbh.Query("select region_name from maps where map_id=$1", id)
	if err != nil {
		return "", fmt.Errorf("Grid DB: Failed to select map ByID. %s", err)
	}
	for rows.Next() {
		var name string

		rows.Scan(&name)

		rows.Close()

		return name, nil
	}
	return "", errors.New("Not found")
}

//RegionTypeByID seek a RegionType by ID
func RegionTypeByID(dbh *db.Handler, id int) (string, error) {

	rows, err := dbh.Query("select region_type from maps where map_id=$1", id)
	if err != nil {
		return "", fmt.Errorf("Grid DB: Failed to select map ByID. %s", err)
	}
	for rows.Next() {
		var name string

		rows.Scan(&name)

		rows.Close()

		return name, nil
	}
	return "", errors.New("Not found")
}

//IDByCityID retrieve grid id by city id.
func IDByCityID(dbh *db.Handler, cityID int) (id int, err error) {
	rows, err := dbh.Query("select map_id from cities where city_id=$1", cityID)
	if err != nil {
		return 0, fmt.Errorf("Grid DB: Failed to select map IDByCityID. %s", err)
	}
	for rows.Next() {
		rows.Scan(&id)
		rows.Close()
		return id, nil
	}

	rows.Close()
	return 0, errors.New("City doesn't exist")
}

//AllShortened seek all grids id and names ;)
func AllShortened(dbh *db.Handler) (grids []*ShortGrid, err error) {
	rows, err := dbh.Exec("select map_id, region_name, region_type, updated_at from maps")
	if err != nil {
		return nil, fmt.Errorf("Grid DB: Failed to select map AllShortened. %s", err)
	}
	for rows.Next() {
		grid := new(ShortGrid)
		rows.Scan(&grid.ID, &grid.Name, &grid.RegionType, &grid.LastUpdate)
		grids = append(grids, grid)
	}

	rows.Close()
	return
}

type dbGrid struct {
	Nodes []node.Node `json:"nodes"`
	Size  int         `json:"size"`
}

func (grid *Grid) dbjsonify() ([]byte, error) {
	var db dbGrid
	db.Nodes = grid.Nodes
	db.Size = grid.Size

	return json.Marshal(db)
}

func (grid *Grid) dbunjsonify(fromJSON []byte) error {
	var db dbGrid
	err := json.Unmarshal(fromJSON, &db)
	if err != nil {
		return err
	}

	grid.Nodes = db.Nodes
	grid.Size = db.Size
	return nil
}
