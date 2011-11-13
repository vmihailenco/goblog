if String::trim is undefined
  String::trim = () ->
    this.replace /^\s+|\s+$/g, ''

String::startswith = (prefix) ->
  return this.indexOf(prefix) == 0

String::endswith = (suffix) ->
  return this.indexOf(suffix, this.length - suffix.length) != -1


mark_active_path = ($container, exact_match=false) ->
  max_length = 0
  $active = null

  $container.find('li').each () ->
    $this = $ @
    href = $this.find('a').attr 'href'

    if href == document.location.pathname
      $active = $this
      return false

    if exact_match
      return

    parts = []
    for part in document.location.pathname.split('/')
      parts.push(part)
      location = parts.join('/')
      location = '/' if location == ''

      if location.length > max_length and
          href.startswith(location)
        max_length = location.length
        $active = $this

  if $active?
    $active.addClass('active')

$ ->
  mark_active_path $('#primary_nav')