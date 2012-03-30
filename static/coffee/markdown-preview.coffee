exports ?= this


exports.preview = ($field, $container, timeout=500) ->
  worker = ->
    text = $field.val()
    $container.text(text)
  worker()

  t = null
  $field.keyup ->
    clearTimeout(t) if t?
    t = setTimeout worker, timeout


exports.previewMarkdown = ($field, $container, timeout=500) ->
  worker = ->
    text = $field.val()

    $.ajax
      url: settings.MARKDOWN_PREVIEW_URL
      type: 'POST'
      data: {text: $field.val()}
      success: (data) -> $container.html(data.html)
  worker()

  t = null
  $field.keyup ->
    clearTimeout(t) if t?
    t = setTimeout worker, timeout
