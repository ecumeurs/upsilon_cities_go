# upsilon_cities_go

# dependencies

* github.com/gorilla/mux : Web router > https://www.gorillatoolkit.org/pkg/mux
* github.com/gorilla/context : Web "request-wide" session > https://www.gorillatoolkit.org/pkg/context
* github.com/gorilla/Sessions : Web session > https://www.gorillatoolkit.org/pkg/Sessions
* github.com/lib/pq : db access > 
* github.com/felixge/httpsnoop : Metrics > https://github.com/felixge/httpsnoop
* github.com/oxtoacart/bpool : template dynamic generation (well it assist)
* github.com/antonlindstrom/pgstore : Web Session (stored in postgresql)

<pre>
# go get github.com/gorilla/mux github.com/gorilla/context github.com/gorilla/Sessions github.com/lib/pq github.com/felixge/httpsnoop github.com/oxtoacart/bpool github.com/antonlindstrom/pgstore
</pre>

# infos 

https://semaphoreci.com/community/tutorials/building-and-testing-a-rest-api-in-go-with-gorilla-mux-and-postgresql 

seem to cover it

# project layout

<pre>
\ config 
\ db
   \ schema.sql
   \ migrations \    # all migrations from 0 to now ;)
\ lib
   \ db \            # DB accessor / Models
   \ cities \        # Cities mechanics
\ web
   \ controllers \   # API/Websever controller 
   \ shared \        # Shared templates (accros all controllers)
   \ layout \        # Main layout ( and specialized layout )
   \ templates \     # Views by controller
   \ static \        # CSS/JS/IMG files
   \ tools \         # Cross purpose functions
   \ router.go       # Router
</pre>

# Config

Don't forget to generate your own config.go !
