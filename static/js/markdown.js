(function() {
  if (typeof exports === "undefined" || exports === null) {
    exports = this;
  }
  exports.markdown = function(text, cb) {
    var html;
    html = '';
    return $.ajax({
      url: settings.MARKDOWN_PREVIEW_URL,
      type: 'POST',
      data: {
        text: text
      },
      success: function(data) {
        return cb(data.html);
      }
    });
  };
}).call(this);
