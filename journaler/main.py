#!/usr/bin/env python3
import os, sys
from datetime import datetime as dt
from pathlib import Path

path = ""

def main():
    args, f = sys.argv[1:], None
    if len(args) == 0: f = new_journal
    elif args[0] == "list": args, f = args[1:], list_journals
    elif args[0] == "read": args, f = args[1:], read_journal
    elif args[0] == "edit": args, f = args[1:], edit_journal
    elif args[0] == "remove": args, f = args[1:], remove_journal
    elif args[0] == "help": f = print_help
    else: print_unknwon_cmd
    f(args)

def new_journal(args):
    t = dt.now()
    t = dt(t.year, t.month, t.day, t.hour, t.minute, 0)
    os.system(f"vim {path / str(int(dt.timestamp(t)))}.txt")

def list_journals(args):
    raw = False
    if len(args) == 1 and args[0] == "raw": raw = True
    elif len(args) != 0: print_unknown_cmd(args)
    for e in get_dir_ents():
        try:
            t = int(e)
            if raw: print(t)
            else: print(dt.fromtimestamp(t).strftime("%H:%M %m/%d/%y"))
        except: pass

def read_journal(args, edit=False, delete=False):
    if len(args) == 1:
        t = 0
        try: t = int(args)
        except:
            try: t = int(dt.timestamp(dt.strptime(args[0], "%H:%M %m/%d/%y")))
            except: print_unknown_cmd(args)
        t = str(t)
        if t in list(get_dir_ents()):
            file_path = f"{path / t}.txt"
            if edit: os.system(f"vim {file_path}")
            elif delete: os.remove(file_path)
            else: os.system(f"view {file_path}")
    else:
        ents = list(get_dir_ents())
        padding_len = len(str(len(ents)))
        for i, e in enumerate(ents):
            t = 0
            try: t = int(e)
            except: continue
            print("{: {}}) {}".format(
                i + 1,
                padding_len,
                dt.fromtimestamp(t).strftime("%H:%M %m/%d/%y")
            ))
        nstr = input("Num (0 to cancel): ")
        n = 0
        try: n = int(nstr) - 1
        except: return
        if n < 0: return
        elif n >= len(ents):
            print("invalid choice")
            return
        file_path = f"{path / ents[n]}.txt"
        if edit: os.system(f"vim {file_path}")
        elif delete: os.remove(file_path)
        else: os.system(f"view {file_path}")

def edit_journal(args): read_journal(args, edit=True)

def remove_journal(args): read_journal(args, delete=True)

def print_help(args):
    print("journaler")
    print("\tlist [raw?]")
    print("\tread [time]")
    print("\tedit [time]")
    print("\tremove [time]")
    print("\thelp")

def print_unknown_cmd(args):
    print(f"unknown command: {' '.join(args)}")
    sys.exit(1)

def get_dir_ents():
    def filt(e):
        print(e.endswith(".txt"))
        return True
    #for p in filter(lambda e: e.endswith(".txt"), os.listdir(path)): print(p)
    try: return map(lambda e: e[:-4], filter(lambda e: e.endswith(".txt"), os.listdir(path)))
    except Exception as e:
        print(f"error reading directory: {e}")
        sys.exit(1)

if __name__ == "__main__":
    path = os.environ.get("JCMDS_PATH")
    if path == "": die('"JCMDS_PATH" environment variable not set')
    path = Path(path) / "journaler"

    main()
