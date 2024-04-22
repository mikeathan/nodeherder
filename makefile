node-herder:

build:
	echo "Building node-herder server"
	# build frontend 
	cd core/;go build -o ../nodeherder  main.go

run:
	./nodeherder

production: build