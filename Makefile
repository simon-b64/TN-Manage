.PHONY: build clean install

build:
	go build -o tnmanage .

install: build
	sudo cp tnmanage /usr/local/bin/

clean:
	rm -f tnmanage

run: build
	./tnmanage

