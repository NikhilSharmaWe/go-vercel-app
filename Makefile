build:
	go build -o ./bin/go-vercel-app

run: build
	./bin/go-vercel-app
