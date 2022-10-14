#!/usr/bin/env python3
import argparse, http.server, socketserver, time, webbrowser
from pathlib import Path
from threading import Thread

class Handler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        directory = str(Path(__file__).parent.parent / "web-pmdr")
        super().__init__(*args, directory=directory, **kwargs)
    def log_message(self, format, *args): pass

parser = argparse.ArgumentParser(description="Run a pomodoro timer on the web")
parser.add_argument("--ip", type=str, default="127.0.0.1",
        help="ip address to listen on (default: 127.0.0.1)", required=False)
parser.add_argument("--port", type=int, default=9090,
        help="port to listen on (default: 9090)", required=False)
parser.add_argument("--no_open", default=False, action="store_true",
        help="open a new browser tab (default: False)", required=False)
args = parser.parse_args()

addr = (args.ip, args.port)

try:
    with socketserver.TCPServer(addr, Handler) as httpd:
        addr = httpd.server_address
        print(f"Running on {addr[0]}:{addr[1]}")
        if not args.no_open:
            Thread(target=webbrowser.open_new_tab, args=(f"http://{addr[0]}:{addr[1]}",)).start()
        try: httpd.serve_forever()
        except KeyboardInterrupt: httpd.shutdown()
except Exception as e:
    print(f"error: {e}")
