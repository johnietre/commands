#!/usr/bin/env python3
aliases = open("/home/johnierodgers/.bashrc")
alias_dict = dict()
for line in aliases.readlines():
    if line.startswith("alias"):
        alias, action = line.replace("alias ", '', 1).split("=", 1)
        alias_dict.setdefault(alias, action.strip())
aliases.close()
aliases = open("/home/johnierodgers/.bash_aliases")
for line in aliases.readlines():
    line = line.strip()
    if line[0] == '#': continue
    alias, action = line.replace("alias ", '', 1).split("=", 1)
    alias_dict.setdefault(alias, action.strip())
aliases.close()
for k, v in sorted(alias_dict.items()):
    print(f"{k}\n\t{v}")

