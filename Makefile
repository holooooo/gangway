build:
	go build -o dist/client ./src/exec/client/main.go
	go build -o dist/controller ./src/exec/controller/main.go
	go build -o dist/repeater ./src/exec/repeater/main.go

package:
	docker build -t gangway/controller .