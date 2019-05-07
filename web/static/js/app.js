
$(document).ready( function() {
    $(".case[data-city]").hover(
        function() {
            
            // on hover, also fetch city related informations and display them in #city_hoder
            $.ajax({
                url: location.href + '/city/' + $(this).data('city'),
                type: 'GET',
                success: function(result) {
                    $('#rightside').html(result)
                }, 
                error: function(result) {
                    
                alert("Failed to drop map... " + result["error"]);
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

});