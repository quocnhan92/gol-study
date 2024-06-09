$(document).ready(function(){
    $('.question-textbox').click(function(){
        var target = $(this).data('target');
        $(target).collapse('toggle');
    });
});