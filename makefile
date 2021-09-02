.PHONY: all aliases pmdr prettypath pwdstore run search sock srvr start uproto

all: aliases pmdr prettypath pwdstore run search sock srvr start uproto

aliases: bin
	cp aliases/main.py bin/aliases && chmod u+x bin/aliases

pmdr: bin
	g++ pmdr/main.cpp -o bin/pmdr -lpthread -std=gnu++17

prettypath: bin
	cp prettypath/main.py bin/prettypath && chmod u+x bin/prettypath

pwdstore: bin
	go build -o bin/pwdstore ./pwdstore

run: bin
	g++ run/main.cpp -o bin/run -std=gnu++17

search: bin
	go build -o bin/search ./search

sock: bin
	go build -o bin/sock ./sock

srvr: bin
	go build -o bin/srvr ./srvr

start: bin
	go build -o bin/start ./start

uproto: bin
	go build -o bin/uproto ./uproto

bin:
	mkdir $@
