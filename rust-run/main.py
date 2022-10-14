import time
"""
print("\x1b[", end='', flush=True)
time.sleep(2)
print("\x1b[5B", end='', flush=True)
print("\x1b 7", end='', flush=True)
print("hello, world", end='', flush=True)
time.sleep(2)
print("\n\n\n", flush=True)
time.sleep(2)
print("\x1b[4A", end='', flush=True)
print("\rgood, world", end='', flush=True)
time.sleep(2)
for i in range(20): print(i)
time.sleep(3)
print("\x1b[15A", end='', flush=True)
print("\x1b[s", end='', flush=True)
time.sleep(3)
print("\x1b[9B", end='', flush=True)
time.sleep(3)
print("\x1b[u", end='', flush=True)
time.sleep(3)
print("\x1b 8", end='', flush=True)
"""
import sys, re
if(sys.platform == "win32"):
    import ctypes
    from ctypes import wintypes
else:
    import termios

def cursorPos():
    if(sys.platform == "win32"):
        OldStdinMode = ctypes.wintypes.DWORD()
        OldStdoutMode = ctypes.wintypes.DWORD()
        kernel32 = ctypes.windll.kernel32
        kernel32.GetConsoleMode(kernel32.GetStdHandle(-10), ctypes.byref(OldStdinMode))
        kernel32.SetConsoleMode(kernel32.GetStdHandle(-10), 0)
        kernel32.GetConsoleMode(kernel32.GetStdHandle(-11), ctypes.byref(OldStdoutMode))
        kernel32.SetConsoleMode(kernel32.GetStdHandle(-11), 7)
    else:
        OldStdinMode = termios.tcgetattr(sys.stdin)
        _ = termios.tcgetattr(sys.stdin)
        _[3] = _[3] & ~(termios.ECHO | termios.ICANON)
        termios.tcsetattr(sys.stdin, termios.TCSAFLUSH, _)
    try:
        _ = ""
        sys.stdout.write("\x1b[6n")
        sys.stdout.flush()
        while not (_ := _ + sys.stdin.read(1)).endswith('R'):
            True
        res = re.match(r".*\[(?P<y>\d*);(?P<x>\d*)R", _)
    finally:
        if(sys.platform == "win32"):
            kernel32.SetConsoleMode(kernel32.GetStdHandle(-10), OldStdinMode)
            kernel32.SetConsoleMode(kernel32.GetStdHandle(-11), OldStdoutMode)
        else:
            termios.tcsetattr(sys.stdin, termios.TCSAFLUSH, OldStdinMode)
    if(res):
        return (res.group("x"), res.group("y"))
    return (-1, -1)

x, y = cursorPos()
print(f"Cursor x: {x}, y: {y}")
