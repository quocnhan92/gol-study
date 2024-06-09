$(document).ready(function () {
    var $navTabs = $('.nav-tabs');
    var tabWidth = $navTabs.find('li').outerWidth(true);
    var visibleTabs = Math.floor($('.nav-tabs-wrapper').width() / tabWidth);
    var totalTabs = $navTabs.find('li').length;
    var currentPosition = 0;

    function shiftTab(direction) {
        var $tabs = $('.nav-tabs > li');
        var $active = $tabs.filter('.active');
        var $newActive;
        if (direction === 'next' && currentPosition < totalTabs - visibleTabs) {
            $newActive = $active.next('li').length ? $active.next('li') : $tabs.first();
            currentPosition++;
        } else if (direction === 'prev' && currentPosition > 0) {
            $newActive = $active.prev('li').length ? $active.prev('li') : $tabs.last();
            currentPosition--;
        }
        $newActive.find('a').tab('show');
        $navTabs.css('transform', 'translateX(' + (-currentPosition * tabWidth) + 'px)');
    }

    $('#next-tab').click(function (e) {
        e.preventDefault();
        shiftTab('next');
    });

    $('#prev-tab').click(function (e) {
        e.preventDefault();
        shiftTab('prev');
    });

            // Show and hide content based on active tab
            $('.nav-tabs > li > a').click(function (e) {
                e.preventDefault();
                var target = $(this).attr('href');
                $('.tab-product-pane').removeClass('tab-active').hide();
                $(target).addClass('tab-active').fadeIn();
            });
    
            // Trigger click on the first tab to show the first content by default
            $('.nav-tabs > li.active > a').trigger('click');
});