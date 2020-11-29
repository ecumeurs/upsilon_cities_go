
reloadCorp = function() {
    console.log("Calling on: " + '/corporation/' + $("#nav-corp").data("corp-id"))
    $.ajax({
        url: '/corporation/' + $("#nav-corp").data("corp-id"),
        type: 'GET',
        success: function(result) {
            
        }, 
        error: function(result) {
            alert(result["error"]);
            window.location.replace("/map");
        }
    });
}

fetchRecentsUserLogs = function() {
    console.log("Calling on: " + '/user/logs/')
    $.ajax({
        url: '/user/logs',
        type: 'GET',
        success: function(result) {
            
        }, 
        error: function(result) {
            alert(result["error"]);
            window.location.replace("/map");
        }
    });
};

corp_reloader_timer = 0
user_logs_timer = 0

$(document).ready( function() {

    $('#CreateUser').on("click",function () {
        forminfo = $("#NewUsrForm").serializeArray()
        $.ajax({
            url: '../api/user/checkavailable/login/' + forminfo[0].value + "/mail/" + forminfo[1].value,
            type: 'GET',
            success: function(result) {
                $('#NewUsrForm').submit()                  
            }, 
            error: function(result) {            
                alert("Le compte n'a pas pu être créé pour les raisons suivantes : " + result.responseJSON["error"]);
            }
        }); 
        }
    );

    $(".container").on("click","#AdminMapButton",function() {
        mapId = $(this).attr("data-map-id")
        myelement = "select[data-map-id='" + mapId + "']"
        corpId = $(myelement).val();
        window.location.href = "/map/" + mapId + "/corp/" + corpId
    });

    $(".action_drop_map").click(function() {
        id = $(this).data("map-id");
        $.ajax({
            url: '/api/map/'+id,
            type: 'DELETE',
            success: function(result) {
                // Do something with the result
                location.reload();
            }, error: function(result) {
                // Do something with the result
                alert("Failed to drop map... " + result["error"]);
                location.reload();
            }
        });
    });

    $(".container").on("click",".href_caravan_action",function() {
        target = $(this).data("target");
        method = $(this).data("method");
        console.log("Attempting to "+method+" to " + target + " reload corp")
        $.ajax({
            url: target,
            type: method,
            success: function(result) {
                // Do forcefully reload corp
                reloadCorp();
            }, error: function(result) {
                // Do something with the result
                alert("Failed to perform request");
                location.reload();
            }
        });
    });

    $(".container").on("click","#city_close",function() {
            $('#city_click').html("")
            $(".city-clicked").removeClass("city-clicked")  
        }
    )
    
    $(".container").on("click",".href_corp_action",function() {
        target = $(this).data("target");
        method = $(this).data("method");
        console.log("Attempting to "+method+" to " + target + " reload corp")
        $.ajax({
            url: target,
            type: method,
            success: function(result) {
                // Do forcefully reload corp
                reloadCorp();
            }, error: function(result) {
                // Do something with the result
                alert("Failed to perform request");
                location.reload();
            }
        });
    });

    $(".container").on("click",".fill_caravan",function() {
        target = $(this).data("target");
        method = $(this).data("method");
        console.log("Attempting to "+method+" to " + target + " and fill caravans.")
        $.ajax({
            url: target,
            type: method,
            success: function(result) {
                // display content in caravan block.
                $("#caravan_holder").html(result)
                
            }, error: function(result) {
                // Do something with the result
                alert("Failed to perform request");
                location.reload();
            }
        });
    });

    
    $(".container").on("click",".href_action",function() {
        target = $(this).data("target");
        method = $(this).data("method");
        redirect = $(this).data("redirect");
        console.log("Attempting to "+method+" to " + target + " and then redirect to " + redirect)
        $.ajax({
            url: target,
            type: method,
            success: function(result) {
                // Do something with the result
                if(redirect == "") {
                    location.reload();
                } else {
                    window.location.replace(redirect)
                }
            }, error: function(result) {
                // Do something with the result
                alert("Failed to perform request");
                location.reload();
            }
        });
    });

    $(".container").on("click",".item_href",function() {
        target = $(this).data("target");
        item = $(this).data("item");
        city_id = $("#city").data("city-id")
        console.log("Attempting to POST to " + "/city/"+city_id+"/"+target+"/"+item )
        $.ajax({
            url: "/city/"+city_id+"/"+target+"/"+item,
            type: "POST",
            success: function(result) {
                // Do something with the result
                $("#item"+item).remove() 
            }, error: function(result) {
                // Do something with the result
                alert("Failed to perform request");
                location.reload();
            }
        });
    });

    $('#city_click').on('click','span.upgrade[data-producer]', function() {
        $('div.upgrade[data-producer=' + $(this).data('producer') + '][data-product=' + $(this).data('product') + ']').toggle()        
    });

    $('#city_click').on('click','span.bigupgrade[data-producer]', function() {
        $('div.bigupgrade[data-producer=' + $(this).data('producer') + '][data-product=' + $(this).data('product') + ']').toggle()
    });

    $('#city_click').on('click','div.upgrade span[data-action]', function() {
        $.ajax({
            url: '/api/city/' + $(this).data('city') + '/producer/' + $(this).data('producer') + '/' + $(this).data('action')+ '/' + $(this).data('product'),
            type: 'POST',
            success: function(result) {
                $.ajax({
                    url: '/city/' + result.CityID,
                    type: 'GET',
                    success: function(result) {
                        $('#city_click').html(result)                 
                    }, 
                    error: function(result) {                        
                        alert("Failed to get city data... " + result["error"]);
                    }
                });              
            }, 
            error: function(result) {                
                alert("Failed to update city data... " + result["error"]);
            }
        });       
    });

    $('#city_click').on('click','div.bigupgrade span[data-action]', function() {
        $.ajax({
            url: '/api/city/' + $(this).data('city') + '/producer/' + $(this).data('producer') + '/' + $(this).data('action')+ '/' + $(this).data('product'),
            type: 'POST',
            success: function(result) {
                $.ajax({
                    url: '/city/' + result.CityID,
                    type: 'GET',
                    success: function(result) {
                        $('#city_click').html(result)                                       
                    }, 
                    error: function(result) {                        
                        alert("Failed to get city data... " + result["error"]);
                    }
                });                
            }, 
            error: function(result) {                
                alert("Failed to update city data... " + result["error"]);
            }
        });       
    });

});