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