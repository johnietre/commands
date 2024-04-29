#!/usr/bin/sh

ghc daylog/Main.hs -o bin/daylog
rm daylog/Main.{hi,o}
