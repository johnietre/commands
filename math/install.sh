#!/usr/bin/sh

cargo +nightly build --release --manifest-path=math/Cargo.toml --bin math && \
  mv math/target/release/math bin/math
