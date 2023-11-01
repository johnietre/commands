#!/usr/bin/env python3
from os import system
from sys import argv

def main():
    args = argv[1:]
    if len(args) == 0 or args[0] in ("-h", "--help"):
        print("Usage: gacp <COMMIT_MESSAGE> [PATHS TO ADD (default -A)]")
        print("Run `git add [PATHS]`, `git commit -m <COMMIT_MESSAGE>`, and `git push` all at once")
        print("NOTE: The message is passed to git commit surrounded by double quotes")
        exit(1)
    conf = input("Did you format code? [Y/n] ").lower()
    if conf not in ("y", "yes"): return

    res = system("git add " + (" ".join(args[1:]) or "-A"))
    if res != 0: exit(res)

    res = system("git commit -m \"" + args[0] + '"')
    if res != 0: exit(res)

    exit(system("git push"))

if __name__ == "__main__":
    try: main()
    except: pass
