$(document).ready(function () {
            // Show and hide content based on active tab
            $('.nav-tabs-hot > li > a').click(function (e) {
                e.preventDefault();
                var target = $(this).attr('href');
                $('.tab-hot-pane').removeClass('tab-active').hide();
                $(target).addClass('tab-active').fadeIn();
            });
    
            // Trigger click on the first tab to show the first content by default
            $('.nav-tabs-hot > li.active > a').trigger('click');
});