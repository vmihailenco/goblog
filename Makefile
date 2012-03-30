compile_css:
	lessc static/less/screen.less static/stylesheets/screen.css
compile_js:
	coffee --compile --watch --output static/js/ static/coffee/
runserver:
	dev_appserver.py .
upload:
	appcfg.py update .
