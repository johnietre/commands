#!/usr/bin/env python3

import os, sys
from pathlib import Path
#from subprocess import run as run_proc, DEVNULL

force = False
cmds_dir = Path(__file__).parent.resolve()

def main():
    global force

    argv = sys.argv[1:]
    if len(argv) == 0:
        print("Usage: make.py [NAMES] [--force]", file=sys.stderr)
        exit(0)

    bin_dir = cmds_dir / "bin"
    if not bin_dir.exists(): bin_dir.mkdir()

    funcs = dict(
        map(
            lambda items: (items[0][11:], items[1]),
            filter(lambda items: items[0].startswith("make_jtcmd_"), globals().items()),
        ),
    )
    funcs["--force"] = noop
    for arg in argv:
        if arg == "--force":
            force = True
            continue
        if funcs.get(arg, None) is None:
            print(f"Unknown argument: {arg}", file=sys.stderr)
            exit(1)
    print_sep = False
    for arg in argv:
        func = funcs[arg]
        if func == noop: continue
        if print_sep: print("="*40)
        print(arg + "\n" + "-"*len(arg))
        func()
        print_sep = True

def noop(): pass

def make_jtcmd_aime():
    if not should_build("bin/aime", ["aime/main.py"]): return
    run_commands(
        Cmd("pip install openai"),
        Cmd("cp aime/main.py bin/aime && chmod u+x bin/aime")
    )

def make_jtcmd_aliases():
    if not should_build("bin/aliases", ["aliases/main.py"]): return
    run_commands(Cmd("cp aliases/main.py bin/aliases && chmod u+x bin/aliases"))

def make_jtcmd_dotodo():
    if not should_build("bin/dotodo", ["dotodo/main.py"]): return
    run_commands(Cmd("cp dotodo/main.py bin/dotodo && chmod u+x bin/dotodo"))

def make_jtcmd_daylog():
    if not should_build("bin/daylog", ["daylog/Main.hs"]): return
    run_commands(
        Cmd("ghc daylog/Main.hs -o bin/daylog"),
        Cmd("rm daylog/Main.o", echo=False),
        Cmd("rm daylog/Main.hi", echo=False),
    )

def make_jtcmd_fmtme():
    if not should_build("bin/fmtme", ["fmtme/main.go"]): return
    run_commands(Cmd("go build -o bin/fmtme ./fmtme"))

def make_jtcmd_finfo():
    if not should_build("bin/finfo", ["finfo/main.go"]): return
    run_commands(Cmd("go build -o bin/finfo ./finfo"))

def make_jtcmd_gacp():
    if not should_build("bin/gacp", ["gacp/main.py"]): return
    run_commands(Cmd("cp gacp/main.py bin/gacp && chmod u+x bin/gacp"))

def make_jtcmd_ghclone():
    if not should_build("bin/ghclone", ["ghclone/main.go"]): return
    run_commands(Cmd("go build -o bin/ghclone ./ghclone"))

def make_jtcmd_journaler():
    if not should_build("bin/journaler", ["journaler/main.py"]): return
    run_commands(Cmd("cp journaler/main.py bin/journaler && chmod u+x bin/journaler"))

def make_jtcmd_linend():
    if not should_build("bin/linend", ["linend/main.rs"]): return
    run_commands(Cmd("rustc linend/main.rs -o bin/linend -C opt-level=3"))

def make_jtcmd_math():
    inputs = ["math/Cargo.toml"] + list((cmds_dir / "math" / "src").glob("*.rs"))
    if not should_build("bin/math", inputs): return
    run_commands(
        Cmd("cargo build --release --manifest-path=math/Cargo.toml"),
        Cmd("mv math/target/release/math bin/math", echo=False),
    )

def make_jtcmd_meyerson():
    inputs = list((cmds_dir / "meyerson").glob("**/*.go"))
    if not should_build("bin/meyerson", inputs): return
    run_commands(Cmd("go build -o bin/meyerson ./meyerson"))

def make_jtcmd_nuid():
    if not should_build("bin/nuid", ["nuid/main.go"]): return
    run_commands(Cmd("go build -o bin/nuid ./nuid"))

def make_jtcmd_pmdr():
    if not should_build("bin/pmdr", ["pmdr/main.cpp"]): return
    run_commands(Cmd("g++ pmdr/main.cpp -Wall -o bin/pmdr -lpthread -std=gnu++17"))

def make_jtcmd_prettypath():
    if not should_build("bin/prettypath", ["prettypath/main.py"]): return
    run_commands(Cmd("cp prettypath/main.py bin/prettypath && chmod u+x bin/prettypath"))

def make_jtcmd_proxyprint():
    if not should_build("bin/proxyprint", ["proxyprint/main.go"]): return
    run_commands(Cmd("go build -o bin/proxyprint ./proxyprint"))

def make_jtcmd_pwdstore():
    if not should_build("bin/pwdstore", ["pwdstore/main.go"]): return
    run_commands(Cmd("go build -o bin/pwdstore ./pwdstore"))

def make_jtcmd_run():
    inputs = ["run/Cargo.toml"] + list((cmds_dir / "run" / "src").glob("*.rs"))
    if not should_build("bin/run", inputs): return
    run_commands(
        Cmd("cargo build --release --manifest-path=run/Cargo.toml"),
        Cmd("mv run/target/release/run bin/run", echo=False),
    )

def make_jtcmd_search():
    inputs = ["search/Cargo.toml"] + list((cmds_dir / "search").glob("*.rs"))
    if not should_build("bin/search", inputs): return
    run_commands(
        Cmd("cargo build --release --manifest-path=search/Cargo.toml"),
        Cmd("mv search/target/release/search bin/search", echo=False),
    )

def make_jtcmd_sock():
    if not should_build("bin/sock", ["sock/main.go"]): return
    run_commands(Cmd("go build -o bin/sock ./sock"))

def make_jtcmd_srvr():
    if not should_build("bin/srvr", ["srvr/main.go"]): return
    run_commands(Cmd("go build -o bin/srvr ./srvr"))

def make_jtcmd_uproto():
    if not should_build("bin/uproto", ["uproto/main.go"]): return
    run_commands(Cmd("go build -o bin/uproto ./uproto"))

def make_jtcmd_vocab():
    if not should_build("bin/vocab", ["vocab/main.go"]): return
    run_commands(Cmd("go build -o bin/vocab ./vocab"))

def make_jtcmd_web_pmdr():
    if not should_build("bin/web-pmdr", ["web-pmdr/main.py"]): return
    run_commands(Cmd("cp web-pmdr/main.py bin/web-pmdr && chmod u+x bin/web-pmdr"))

class Cmd:
    def __init__(self, cmd, echo=True, ignore_err=False):
        self.cmd = cmd
        self.echo = echo
        self.ignore_err = ignore_err

    def run(self):
        if self.echo: print(self.cmd)
        #return run_proc(self.cmd).returncode
        return os.system(self.cmd)

def run_commands(*cmds):
    for cmd in cmds:
        code = cmd.run()
        if code != 0 and cmd.ignore_err: exit(code)

def should_build(output, inputs):
    global force
    if force: return True
    output = Path(output)
    if not output.exists(): return True
    omt = output.stat().st_mtime
    return any(map(lambda inp: Path(inp).stat().st_mtime > omt, inputs))

if __name__ == "__main__":
    main()
