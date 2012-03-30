(function() {
  if (typeof exports === "undefined" || exports === null) {
    exports = this;
  }
  exports.preview = function($field, $container, timeout) {
    var t, worker;
    if (timeout == null) {
      timeout = 500;
    }
    worker = function() {
      var text;
      text = $field.val();
      return $container.text(text);
    };
    worker();
    t = null;
    return $field.keyup(function() {
      if (t != null) {
        clearTimeout(t);
      }
      return t = setTimeout(worker, timeout);
    });
  };
  exports.previewMarkdown = function($field, $container, timeout) {
    var t, worker;
    if (timeout == null) {
      timeout = 500;
    }
    worker = function() {
      var text;
      text = $field.val();
      return $.ajax({
        url: settings.MARKDOWN_PREVIEW_URL,
        type: 'POST',
        data: {
          text: $field.val()
        },
        success: function(data) {
          return $container.html(data.html);
        }
      });
    };
    worker();
    t = null;
    return $field.keyup(function() {
      if (t != null) {
        clearTimeout(t);
      }
      return t = setTimeout(worker, timeout);
    });
  };
}).call(this);
