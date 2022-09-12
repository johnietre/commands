.PHONY: all aliases articles daylog dotodo journaler math meyerson pmdr prettypath pwdstore run search sock srvr start uproto vocab web-pmdr

all: aliases daylog journaler math meyerson pmdr prettypath pwdstore run search sock srvr start uproto vocab web-pmdr

aliases: bin
	cp aliases/main.py bin/aliases && chmod u+x bin/aliases

articles: bin
	go build -o bin/articles ./articles

dotodo: bin
	cp dotodo/main.py bin/dotodo && chmod u+x bin/dotodo

daylog: bin
	ghc daylog/Main.hs -o bin/daylog
	@rm daylog/Main.o
	@rm daylog/Main.hi

journaler: bin
	cp journaler/main.py bin/journaler && chmod u+x bin/journaler

math: bin
	cd math && cargo build --release && mv target/release/math ../bin/math

meyerson: bin
	go build -o bin/meyerson ./meyerson

pmdr: bin
	g++ pmdr/main.cpp -Wall -o bin/pmdr -lpthread -std=gnu++17

prettypath: bin
	cp prettypath/main.py bin/prettypath && chmod u+x bin/prettypath

pwdstore: bin
	go build -o bin/pwdstore ./pwdstore

run: bin
	#g++ run/main.cpp -Wall -o bin/run -std=gnu++17
	cd rust-run && cargo build --release && mv target/release/rust-run ../bin/run

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

vocab: bin
	go build -o bin/vocab ./vocab

web-pmdr: bin
	cp web-pmdr/main.py bin/web-pmdr && chmod u+x bin/web-pmdr

bin:
	mkdir $@
