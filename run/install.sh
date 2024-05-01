#!/usr/bin/sh

#g++ run/cpp/main.cpp -Wall -o bin/run -std=gnu++17
cargo build --release --manifest-path=run/Cargo.toml --bin run && \
  mv run/target/release/run bin/run
