.PHONY: dev

dev:
	bash livereload.sh

build:
	templ generate
	tailwindcss -i ./public/css/base.css -o ./public/css/tailwind.css
	go build -o zapback main.go
