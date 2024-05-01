#!/usr/bin/sh

cd math && cargo build --release && mv target/release/math bin/math
