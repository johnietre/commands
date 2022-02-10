#!/usr/bin/env python3
import argparse, http.server, socketserver
from pathlib import Path

class Handler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        directory = str(Path(__file__).parent.parent / "web-pmdr")
        super().__init__(*args, directory=directory, **kwargs)
    def log_message(self, format, *args): pass

parser = argparse.ArgumentParser(description="Run a pomodoro timer on the web")
parser.add_argument("--ip", type=str, default="",
        help="ip address to listen on", required=False)
parser.add_argument("--port", type=int, default=9090,
        help="port to listen on (default: 9090)", required=False)
args = parser.parse_args()

addr = (args.ip, args.port)

with socketserver.TCPServer(addr, Handler) as httpd:
    addr = httpd.server_address
    print(f"Running on {addr[0]}:{addr[1]}")
    try: httpd.serve_forever()
    except KeyboardInterrupt: httpd.shutdown()
    except Exception as e: print(f"error: {e}")
