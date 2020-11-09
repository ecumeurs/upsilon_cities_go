/**
 * Author: Michael Hadley, mikewesthad.com
 * Asset Credits:
 *  - Nintendo Mario Tiles, Educational use only
 */

const config = {
  type: Phaser.AUTO,
  width: 320,
  height: 320,
  zoom: 1, // Since we're working with 16x16 pixel tiles, let's scale up the canvas by 4x
  pixelArt: true, // Force the game to scale images up crisply
  parent: "game-container",
  scene: {
    preload: preload,
    create: create
  }
};

const game = new Phaser.Game(config);

function preload() {
  this.load.image("tiles", "/static/assets/tilesets/OverWorld.png");
}

function create() {
  // Load a map from a 2D array of tile indices
  // prettier-ignore
  var gamescene = this

  $.ajax({
    url: '../api/map/3',
    type: 'GET',
    success: function(result) {  

      var emptymap = [
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ],
        [1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1,  1  ]
      ]


      const map = gamescene.make.tilemap({ data:emptymap, tileWidth: 16, tileHeight: 16 });
      const tiles = map.addTilesetImage("tiles"); 

      map.currentLayerIndex = 0;
      const groundmap = map.createBlankDynamicLayer('groundmap', tiles);
      map.currentLayerIndex = 1;
      const envmap = map.createBlankDynamicLayer('envmap', tiles);  
      map.currentLayerIndex = 2;
      const roadmap = map.createBlankDynamicLayer('roadmap', tiles); 
      map.currentLayerIndex = 3;
      const structmap = map.createBlankDynamicLayer('structmap', tiles);  

      result.WebGrid.Nodes.forEach(function(array){
        array.forEach(function(item){

          console.log(item.Node);
          if(item.Node.IsStructure)
          {
            structmap.putTileAt(43,item.Node.Location.X,item.Node.Location.Y);
          }
          else if(item.Node.IsRoad )
          {
            roadmap.putTileAt(9,item.Node.Location.X,item.Node.Location.Y);
          }
          
          if( item.Node.Landscape != 0)
          {
            envmap.putTileAt(item.Node.Landscape,item.Node.Location.X,item.Node.Location.Y);
          }
          
          if( item.Node.Ground != 0)
          {
            groundmap.putTileAt(item.Node.Ground,item.Node.Location.X,item.Node.Location.Y);
          }
        })
      });

    }, 
    error: function(result) {        
      alert("Failed to get city data... " + result["error"]);
    }
}); 

 

}