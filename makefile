.PHONY: all aliases pmdr prettypath pwdstore run search sock srvr start uproto

all: aliases pmdr prettypath pwdstore run search sock srvr start uproto

aliases:
	cp aliases/main.py bin/aliases && chmod u+x bin/aliases

pmdr:
	g++ pmdr/main.cpp -o bin/pmdr -lpthread -std=gnu++17

prettypath:
	cp prettypath/main.py bin/prettypath && chmod u+x bin/prettypath

pwdstore:
	go build -o bin/pwdstore pwdstore/main.go

run:
	g++ run/main.cpp -o bin/run -std=gnu++17

search:
	go build -o bin/search search/main.go

sock:
	go build -o bin/sock ./sock

srvr:
	go build -o bin/srvr srvr/main.go

start:
	go build -o bin/start start/main.go

uproto:
	go build -o bin/uproto uproto/main.go
