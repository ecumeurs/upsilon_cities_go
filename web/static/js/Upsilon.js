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
      'url': "../static/json/tileset.json",
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
    url: '../api' + window.location.pathname,
    type: 'GET',
      'success': function (result) {
          mapTmp = result;
      }, 
      error: function(result) {        
        alert("Failed to get city data... " + result["error"]);
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
    create: create
  }
};

const game = new Phaser.Game(config);

function preload() {
  this.load.image("tiles", "/static/assets/tilesets/OverWorld.png");
}

function getTile(node,type,table){

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


  const map = gamescene.make.tilemap({ data:emptymap, tileWidth: 32, tileHeight: 32 });
  const tiles = map.addTilesetImage("tiles"); 

  map.currentLayerIndex = 0;
  const groundmap = map.createBlankDynamicLayer('groundmap', tiles);
  map.currentLayerIndex = 1;
  const envmap = map.createBlankDynamicLayer('envmap', tiles);  
  map.currentLayerIndex = 2;
  const roadmap = map.createBlankDynamicLayer('roadmap', tiles); 
  map.currentLayerIndex = 3;
  const structmap = map.createBlankDynamicLayer('structmap', tiles); 
  
  mapInfo.WebGrid.Nodes.forEach(function(array){
    array.forEach(function(item){          
      if(item.Node.IsStructure)
      {      
        structmap.putTileAt((getTile(item.Node,"Structure",table)),item.Node.Location.X,item.Node.Location.Y);
      }
        
      if(item.Node.IsRoad )
      {
        roadmap.putTileAt((getTile(item.Node,"Road",table)),item.Node.Location.X,item.Node.Location.Y);
      }
      
      if( item.Node.Landscape != "NoLandscape")
      {            
        envmap.putTileAt((getTile(item.Node,"Landscape",table)),item.Node.Location.X,item.Node.Location.Y);
      }
      
      if( item.Node.Ground != "NoGround")
      {
        groundmap.putTileAt((getTile(item.Node,"Ground",table)),item.Node.Location.X,item.Node.Location.Y);
      }

    })
  });  

  this.input.on(Phaser.Input.Events.POINTER_DOWN, (pointer) => {
    
    console.log(window.location.pathname)

    var tileworldX = pointer.worldX - (pointer.worldX%16);    
    var tileworldY = pointer.worldY - (pointer.worldY%16);    
    //var tileX = pointer.worldX / tileWidth;    
    //var tileY = pointer.worldY / tileHeight;    
    
    const targetVec =  groundmap.worldToTileXY(tileworldX, tileworldY)
    console.log(targetVec)
    $(".city-clicked").removeClass("city-clicked")
    $(this).toggleClass("city-clicked"); 

    $("#city_click").removeClass("city-menu-click")
    $.ajax({
        url: '../api' + window.location.pathname + '/city/X/' + targetVec.x + "/Y/" + targetVec.y,
        type: 'GET',
        success: function(result) {
          console.log(result)
            $('#city_click').html(result)                   
        }, 
        error: function(result) {
            
        alert("Failed to get city data... " + result["error"]);
        }
    }); 
    $("#city_click").toggleClass("city-menu-click")
    
  })
  
  $(".case[data-city]").click(        
    function() {
        // on hover, also fetch city related informations and display them in #city_hoder
 
    }
  )
}