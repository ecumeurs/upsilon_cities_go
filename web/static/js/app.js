
$(document).ready( function() {
    $(".case[data-city]").hover(
        function() {
            $(this).toggleClass("city-hovered");
            neighbours = eval($(this).attr("data-neighbours"))
            console.log(neighbours)
            neighbours.forEach(element => {
                $(".case[data-loc='"+element+"']").toggleClass("target-city-hovered")
            });
        },
        function() {
            $(this).toggleClass("city-hovered");   
            neighbours = eval($(this).attr("data-neighbours"))
            neighbours.forEach(element => {
                $(".case[data-loc='"+element+"']").toggleClass("target-city-hovered")
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