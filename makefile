node-herder:

build_backend:
	echo "Building node-herder backend"
	rm ./nodeherder; cd core/;  go build -o ../nodeherder  main.go

build_frontend:
	echo "Building node-herder frontend"
	cd frontend/; rm -fe dist/; npm run build

run:
	./nodeherder

build: build_backend build_frontend