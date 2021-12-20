#!/usr/bin/env python3
aliases = open("/home/johnierodgers/.bashrc")
alias_dict = dict()
prev = ""
for line in aliases.readlines():
    if line.startswith("alias"):
        alias, action = line.replace("alias ", '', 1).split("=", 1)
        alias_dict.setdefault(alias, action.strip())
    elif "()" in line:
        func = ''.join(filter(lambda c: c.isalpha(), line))
        desc = "No description"
        if prev and prev[0] == '#': desc = prev[1:].strip()
        alias_dict.setdefault(func, desc)
    prev = line
aliases.close()
aliases = open("/home/johnierodgers/.bash_aliases")
for line in aliases.readlines():
    line = line.strip()
    if not line or line[0] == '#': continue
    alias, action = line.replace("alias ", '', 1).split("=", 1)
    alias_dict.setdefault(alias, action.strip())
aliases.close()
for k, v in sorted(alias_dict.items()):
    print(f"{k}\n\t{v}")

