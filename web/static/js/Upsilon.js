/**
 * Author: Michael Hadley, mikewesthad.com
 * Asset Credits:
 *  - Nintendo Mario Tiles, Educational use only
 */

const table = (function () {
  var json = null;
  $.ajax({
      'async': false,
      'global': false,
      'url': "/static/json/tileset.json",
      'dataType': "json",
      'success': function (data) {
          json = data;
      }
  });
  return json;
})(); 

const mapInfo = (function () {
  var mapTmp = null;
  $.ajax({    
    'async': false,
    'global': false,
    url: "/api/" + window.location.pathname,
    type: 'GET',
      'success': function (result) {
          mapTmp = result;
      }, 
      error: function(result) {        
        alert("(Phaser3) Failed to get city data... " + result["error"]);
      }
  });
  return mapTmp;
})(); 

const minWidth = mapInfo.WebGrid.Nodes.length*32
const minHeigh =  mapInfo.WebGrid.Nodes.length*32

const config = {
  type: Phaser.AUTO,
  scale: {
    parent: 'game-container',
    mode: Phaser.Scale.FIT,
    autoCenter: Phaser.Scale.CENTER_HORIZONTALLY,
    width: minWidth,
    height: minHeigh,
    max :{ 
    width: minWidth,
    height: minHeigh
    }
  }, // Force the game to scale images up crisply 
  scene: {
    preload: preload,
    create: create,
    update: update
  }
};

const game = new Phaser.Game(config);

var marker;
var layer;
var map;
var groundmap;
var envmap;
var roadmap;
var structmap;

function preload() {
  this.load.image("tiles", "/static/assets/tilesets/OverWorld.png");
}

function getTileInfo(node,type,table){

  var myInt = 0; 

  switch(type)
  {
    case 'Landscape' :
      myInt = table[type][node.Landscape]
      break;
    case 'Ground' :
      myInt = table[type][node.Ground]
      break;
    case 'Structure' :
      myInt = table[type]["City"] //table[type][node.Structure]
      break;
    case 'Road' :
      myInt = table[type]["Road"] //table[type][node.Road]
      break;
  }

  return myInt
}

function create() {
  // Load a map from a 2D array of tile indices
  // prettier-ignore
  var gamescene = this  
  
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


  map = gamescene.make.tilemap({ data:emptymap, tileWidth: 32, tileHeight: 32 });
  var tiles = map.addTilesetImage("tiles"); 

  map.currentLayerIndex = 0;
  groundmap = map.createBlankDynamicLayer('groundmap', tiles);
  map.currentLayerIndex = 1;
  envmap = map.createBlankDynamicLayer('envmap', tiles);  
  map.currentLayerIndex = 2;
  roadmap = map.createBlankDynamicLayer('roadmap', tiles); 
  map.currentLayerIndex = 3;
  structmap = map.createBlankDynamicLayer('structmap', tiles); 
  
  mapInfo.WebGrid.Nodes.forEach(function(array){
    array.forEach(function(item){    
        if(item.Node.IsStructure)
        {      
          structmap.putTileAt((getTileInfo(item.Node,"Structure",table)),item.Node.Location.X,item.Node.Location.Y);
        }
          
        if(item.Node.IsRoad )
        {
          roadmap.putTileAt((getTileInfo(item.Node,"Road",table)),item.Node.Location.X,item.Node.Location.Y);
        }
        
        if( item.Node.Landscape != "NoLandscape")
        {            
          envmap.putTileAt((getTileInfo(item.Node,"Landscape",table)),item.Node.Location.X,item.Node.Location.Y);
        }
        
        if( item.Node.Ground != "NoGround")
        {
          groundmap.putTileAt((getTileInfo(item.Node,"Ground",table)),item.Node.Location.X,item.Node.Location.Y);
        }
    })

  });  

  marker = this.add.graphics();
  marker.lineStyle(1, 0x000000);
  marker.strokeRect(0, 0, 32, 32);

  this.input.on(Phaser.Input.Events.POINTER_DOWN, (pointer) => {
    
    var tileworldX = pointer.worldX - (pointer.worldX%32);    
    var tileworldY = pointer.worldY - (pointer.worldY%32);      
   
    var myVec =  groundmap.worldToTileXY(tileworldX, tileworldY);
    var tile = mapInfo.WebGrid.Nodes[myVec.y][myVec.x]
    if( tile.Node.IsStructure ){   
    
      $(".city-clicked").removeClass("city-clicked");
      $(this).toggleClass("city-clicked"); 
    
      $("#city_click").removeClass("city-menu-click");

      $.ajax({
          url: '/api' + window.location.pathname + '/city/X/' + myVec.x + "/Y/" + myVec.y,
          type: 'GET',
          success: function(result) {
              $('#city_click').html(result);
          }, 
          error: function(result) {            
            alert("Failed to get city data... " + result["error"]);
          }
      }); 

      $("#city_click").toggleClass("city-menu-click");

    }
  })

}

function update() {

  var pointer = game.input.activePointer;
  var tileworldX = pointer.worldX - (pointer.worldX%32);
  var tileworldY = pointer.worldY - (pointer.worldY%32);
  var myVec =  groundmap.worldToTileXY(tileworldX, tileworldY);

  if( marker.x != myVec.x * 32 || marker.y != myVec.y * 32){
    marker.x = myVec.x * 32;
    marker.y = myVec.y * 32;
    var el = mapInfo.WebGrid.Nodes[myVec.y][myVec.x]
    var tileInfo = "";

    if( el.Node.IsRoad ){
      tileInfo += "Route ";
    }

    if( el.Node.IsStructure ){
      tileInfo += "Ville ";
    }

    if( el.Node.Ground != "NoGround" )
    {
      tileInfo += el.Node.Ground + " ";
    }

    if( el.Node.Landscape != "NoLandscape" )
    {
      tileInfo += el.Node.Landscape + " ";
    }

    $("#TileInfo").text(tileInfo);
    $("#TileInfoX").text(myVec.x);
    $("#TileInfoY").text(myVec.y);

  }

}