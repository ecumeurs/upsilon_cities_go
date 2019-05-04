package grid

import (
	"encoding/json"
	"errors"
	"log"
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

	rows := dbh.Query("insert into maps(region_name, data) values($1,$2) returning map_id", grid.Name, json)
	for rows.Next() {
		rows.Scan(&grid.ID)
	}

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

	dbh.Query(`update maps set
			region_name=$1,
			data=$2,
			updated_at= (now() at time zone 'utc') 
			where map_id=$3;`, grid.Name, json, grid.ID)
	log.Printf("Grid: Grid %d Updated", grid.ID)
	return nil
}

//Drop grid from database
func (grid *Grid) Drop(dbh *db.Handler) error {
	dbh.Query("delete from maps where map_id=$1", grid.ID)
	log.Printf("Grid: Grid %d Deleted", grid.ID)
	grid.ID = 0

	return nil
}

//ByID seek a grid by ID
func ByID(dbh *db.Handler, id int) (grid *Grid, err error) {
	rows := dbh.Query("select map_id, region_name, updated_at, data from maps where map_id=$1", id)
	for rows.Next() {
		grid := new(Grid)
		var json []byte
		rows.Scan(&grid.ID, &grid.Name, &grid.LastUpdate, &json)
		grid.dbunjsonify(json)
		return grid, nil
	}
	return nil, errors.New("Not found")
}

//AllShortened seek all grids id and names ;)
func AllShortened(dbh *db.Handler) (grids []*ShortGrid, err error) {
	rows := dbh.Exec("select map_id, region_name, updated_at from maps")
	for rows.Next() {
		grid := new(ShortGrid)
		rows.Scan(&grid.ID, &grid.Name, &grid.LastUpdate)
		grids = append(grids, grid)
	}
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
