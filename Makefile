
build:
	go build -o bin/wow-fisher cmd/wow-fisher.go

run: build
	bin/wow-fisher