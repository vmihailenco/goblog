exports ?= this


exports.markdown = (text, cb) ->
  html = ''
  $.ajax
    url: settings.MARKDOWN_PREVIEW_URL
    type: 'POST'
    data: {text: text}
    success: (data) -> cb data.html
