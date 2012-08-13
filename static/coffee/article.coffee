do ->
  delay = null
  timeout = 1000

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
    markdown editor.getValue(), (html) -> $preview.html html

  delay = setTimeout update, timeout


do ->
  delay = null
  timeout = 1000

  $title = $('#Title')
  $preview = $('#titleHTML')

  $title.keyup ->
    clearTimeout delay
    delay = setTimeout update, timeout

  update = ->
    $preview.text $title.val()

  delay = setTimeout update, timeout


do ->
  $('#Image').fileupload
    submit: (e, data) ->
      $this = $ this
      $.ajax
        url: settings.IMAGE_UPLOAD_URL
        success: (result) ->
          data.url = result.url
          $this.fileupload 'send', data
      false
