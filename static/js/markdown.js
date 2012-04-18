(function() {
  if (typeof exports === "undefined" || exports === null) {
    exports = this;
  }
  exports.markdown = function(text) {
    var html;
    html = '';
    $.ajax({
      async: false,
      url: settings.MARKDOWN_PREVIEW_URL,
      type: 'POST',
      data: {
        text: text
      },
      success: function(data) {
        return html = data.html;
      }
    });
    return html;
  };
}).call(this);
