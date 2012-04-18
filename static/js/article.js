(function() {
  (function() {
    var $preview, delay, editor, timeout, update;
    delay = null;
    timeout = 500;
    $preview = $('#textHTML');
    editor = CodeMirror.fromTextArea(document.getElementById('Text'), {
      mode: 'markdown',
      lineNumbers: true,
      matchBrackets: true,
      theme: 'default',
      onChange: function() {
        clearTimeout(delay);
        return delay = setTimeout(update, timeout);
      }
    });
    update = function() {
      return $preview.html(markdown(editor.getValue()));
    };
    return delay = setTimeout(update, timeout);
  })();
  (function() {
    var $preview, $title, delay, timeout, update;
    delay = null;
    timeout = 500;
    $title = $('#Title');
    $preview = $('#titleHTML');
    $title.keyup(function() {
      clearTimeout(delay);
      return delay = setTimeout(update, timeout);
    });
    update = function() {
      return $preview.text($title.val());
    };
    return delay = setTimeout(update, timeout);
  })();
}).call(this);
