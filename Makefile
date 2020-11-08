.PHONY: test install clean coverage run_as_service remove_service all compile

all: compile run_as_service

install:
	go get github.com/mattn/goveralls
	go get -u github.com/rakyll/gotest

test:
	gotest -v ./...

coverage:
	gotest -covermode=count -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf coverage.out
	rm -rf coverage.html

run_as_service:
	sudo cp solargo.service /lib/systemd/system/solargo.service
	sudo chmod 644 /lib/systemd/system/solargo.service
	sudo systemctl enable solargo
	sudo service solargo start

remove_service:
	sudo systemctl stop solargo
	sudo systemctl disable solargo
	sudo rm /lib/systemd/system/solargo.service

compile:
	go build -ldflags "-s -w"

