#!/usr/bin/env python3
from os import getenv
for p in sorted(getenv("PATH").split(":")): print(p)
