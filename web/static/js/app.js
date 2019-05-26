
reloadCorp = function() {
    console.log("Calling on: " + '/corporation/' + $("#corp").data("corp-id"))
    $.ajax({
        url: '/corporation/' + $("#corp").data("corp-id"),
        type: 'GET',
        success: function(result) {
            $('#corp').html(result)                   
        }, 
        error: function(result) {
            
        alert("Failed to get corporation data... " + result["error"]);
        }
    });
}

corp_reloader_timer = 0

$(document).ready( function() {

    reloadCorp()
    corp_reloader_timer = setInterval(reloadCorp, 5000);

    $(".case[data-city]").hover(
        function() {
            // on hover, also fetch city related informations and display them in #city_hoder
            $.ajax({
                url: '/city/' + $(this).data('city'),
                type: 'GET',
                success: function(result) {
                    $('#rightside').html(result)                   
                }, 
                error: function(result) {
                    
                alert("Failed to get city data... " + result["error"]);
                }
            });
            
            console.log("hovering cities ...")
            $(this).toggleClass("city-hovered");
            neighbours = eval($(this).attr("data-neighbours"))
            console.log(neighbours)
            neighbours.forEach(element => {
                $(".case[data-loc='"+element+"']").addClass("target-city-hovered")
                console.log("hovering cities ... "  + element )
            });

        },
        function() {
            $(this).toggleClass("city-hovered");   
            neighbours = eval($(this).attr("data-neighbours"))
            console.log("dehovering cities ...")
            neighbours.forEach(element => {
                console.log("dehovering cities ... " + element )
                $(".case[data-loc='"+element+"']").removeClass("target-city-hovered")
            });      
        }
    )

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

    $("#rightside").on("click",".href_caravan_action",function() {
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

    $("#rightside").on("click",".fill_caravan",function() {
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

    
    $("#rightside").on("click",".href_action",function() {
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

    $('#rightside').on('click','span.upgrade[data-producer]', function() {
        $('div.upgrade[data-producer=' + $(this).data('producer') + ']').toggle()        
    });

    $('#rightside').on('click','span.bigupgrade[data-producer]', function() {
        $('div.bigupgrade[data-producer=' + $(this).data('producer') + ']').toggle()
    });

    $('#rightside').on('click','div.upgrade span[data-action]', function() {
        $.ajax({
            url: '/api/city/' + $(this).data('city') + '/producer/' + $(this).data('producer') + '/' + $(this).data('action')+ '/' + $(this).data('product'),
            type: 'POST',
            success: function(result) {
                $.ajax({
                    url: '/city/' + result.CityID,
                    type: 'GET',
                    success: function(result) {
                        $('#rightside').html(result)                 
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

    $('#rightside').on('click','div.bigupgrade span[data-action]', function() {
        $.ajax({
            url: '/api/city/' + $(this).data('city') + '/producer/' + $(this).data('producer') + '/' + $(this).data('action')+ '/' + $(this).data('product'),
            type: 'POST',
            success: function(result) {
                $.ajax({
                    url: '/city/' + result.CityID,
                    type: 'GET',
                    success: function(result) {
                        $('#rightside').html(result)                                       
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