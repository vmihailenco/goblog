exports ?= this


exports.markdown = (text) ->
  html = ''
  $.ajax
    async: false
    url: settings.MARKDOWN_PREVIEW_URL
    type: 'POST'
    data: {text: text}
    success: (data) -> html = data.html
  return html
