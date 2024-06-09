$(document).ready(function() {
    var suggestions = [
        "How to send money?",
        "How to withdraw money?",
        "How to check my balance?",
        "How to update my profile?",
        "How to contact support?"
    ];

    $('#searchBox').on('input', function() {
        var query = $(this).val().toLowerCase();
        var matches = suggestions.filter(function(item) {
            return item.toLowerCase().includes(query);
        });

        var suggestionsList = $('#suggestionsList');
        suggestionsList.empty();
        if (matches.length > 0 && query.length > 0) {
            matches.forEach(function(item) {
                suggestionsList.append('<li>' + item + '</li>');
            });
            suggestionsList.show();
        } else {
            suggestionsList.hide();
        }
    });

    $('#suggestionsList').on('click', 'li', function() {
        $('#searchBox').val($(this).text());
        $('#suggestionsList').hide();
    });

    $('#searchButton').click(function() {
        alert('Search for: ' + $('#searchBox').val());
    });

    $(document).click(function(e) {
        if (!$(e.target).closest('.search-container').length) {
            $('#suggestionsList').hide();
        }
    });
});