$(function() {
    var video = $("#video");
    
    $(video).on("seeking", function () {
        // var currentTime = $(video)[0].currentTime;
        // document.cookie = "played_time=" + currentTime;
        // console.log("seeking at " + currentTime)
    });

    $(video).on("progress", function () {
        var currentTime = $(video)[0].currentTime;
        document.cookie = "played_time=" + currentTime;
        // document.cookie = "progress_at=" + currentTime;
        // console.log("progressing at " + currentTime)
    });
});