all: resources.go

resources.go: resources/*
	go-bindata --prefix resources -pkg data -o $@ resources/...

reset:
	ln -s 
