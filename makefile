.PHONY: all aime aliases articles daylog dotodo finfo fmtme gacp ghclone journaler linend math meyerson nuid pmdr prettypath proxyprint pwdstore run search sock srvr start uproto vocab web-pmdr

all: aime aliases articles daylog dotodo fmtme finfo gacp ghclone journaler linend math meyerson nuid pmdr prettypath proxyprint pwdstore run search sock srvr start uproto vocab web-pmdr

aime: bin
	pip install openai
	cp aime/main.py bin/aime && chmod u+x bin/aime

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

fmtme: bin
	go build -o bin/fmtme ./fmtme

finfo: bin
	go build -o bin/finfo ./finfo

gacp: bin
	cp gacp/main.py bin/gacp && chmod u+x bin/gacp

ghclone: bin
	go build -o bin/ghclone ./ghclone

journaler: bin
	cp journaler/main.py bin/journaler && chmod u+x bin/journaler

linend: bin
	rustc linend/main.rs -o bin/linend -C opt-level=3

math: bin
	cd math && cargo build --release && mv target/release/math ../bin/math

meyerson: bin
	go build -o bin/meyerson ./meyerson

nuid: bin
	go build -o bin/nuid ./nuid

pmdr: bin
	g++ pmdr/main.cpp -Wall -o bin/pmdr -lpthread -std=gnu++17

prettypath: bin
	cp prettypath/main.py bin/prettypath && chmod u+x bin/prettypath

proxyprint: bin
	go build -o bin/proxyprint ./proxyprint

pwdstore: bin
	go build -o bin/pwdstore ./pwdstore

run: bin
	#g++ run/cpp/main.cpp -Wall -o bin/run -std=gnu++17
	cargo build --release --manifest-path=run/Cargo.toml --bin run && \
		mv run/target/release/run ../bin/run

search: bin
	#go build -o bin/search ./search
	cargo build --release --manifest-path=search/Cargo.toml && \
		mv search/target/release/search ../bin/search

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
