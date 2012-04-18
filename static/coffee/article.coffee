do ->
  delay = null
  timeout = 500

  $preview = $('#textHTML')

  editor = CodeMirror.fromTextArea document.getElementById('Text'),
    mode: 'markdown'
    lineNumbers: true
    lineWrapping: true
    matchBrackets: true
    theme: 'default'
    onChange: () ->
      clearTimeout delay
      delay = setTimeout update, timeout

  update = ->
    $preview.html markdown(editor.getValue())

  delay = setTimeout update, timeout


do ->
  delay = null
  timeout = 500

  $title = $('#Title')
  $preview = $('#titleHTML')

  $title.keyup ->
    clearTimeout delay
    delay = setTimeout update, timeout

  update = ->
    $preview.text $title.val()

  delay = setTimeout update, timeout
