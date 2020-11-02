// Copy this file to config/config.go
// Set appropriate values for consts.
package config

import "os"

const (
    DB_USER     = "postgres"
    DB_PASSWORD = "Azerty123!"
    DB_NAME     = "upsilon_cities_go"
    TEST_DB_NAME     = "upsilon_cities_go_test"
    DB_HOST     = "127.0.0.1"
	DB_PORT       = "5432"
    HTTP_PORT   = ":80"
    STATIC_FILES  = "web/static"
	WEB_TEMPLATES = "web/templates"
	WEB_LAYOUTS   = "web/layouts"
	WEB_SHARED    = "web/shared"
	DB_MIGRATIONS = "db/migrations"
	DB_SCHEMA     = "db/schema.sql"
	DB_SEEDS     = "db/seeds"
	WEB_RELOADING = true
	DATA_PRODUCERS = "data/producers"
	DATA_NAMES     = "data/names"
	SYS_DIR_SEP 	 = string(os.PathSeparator)
	SYS_FORCE_ROOT = false
	SYS_ROOT       = "/"
	SESSION_SECRET_KEY  = "aziejaoifndngroeigbgfdkjsgs45435"
	USER_ENABLED_BY_DEFAULT = true
	USER_ADMIN_BY_DEFAULT = true 
	FAME_LOSS_BY_SPACE = -1
	FAME_GAIN_BY_CARAVAN    = 20
	FAME_LOSS_BY_CARAVAN    = -50
	INITIAL_CITY_STORAGE_SIZE = 500
	PRODUCABLE_ITEM_PRICE = 0.5
	UNPRODUCABLE_ITEM_PRICE = 1
	PRODUCABLE_ITEM_FAME = 0.1
	UNPRODUCABLE_ITEM_FAME = 0.5

)
