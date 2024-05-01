#!/usr/bin/sh

#go build -o bin/search ./search
cargo build --release --manifest-path=search/Cargo.toml && \
  mv search/target/release/search bin/search
