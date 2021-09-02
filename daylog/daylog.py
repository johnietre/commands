#!/usr/bin/env python3.9
from sqlite3 import connect
from sys import version_info, exit
from time import time
from datetime import datetime
from os import path

hour_to_check = 0

if version_info[1] < 9: dir_path = path.dirname(path.realpath(__file__))
else: dir_path = path.dirname(__file__)

conn = connect(path.join(dir_path, "daily_log.db"))
c = conn.cursor()

last = c.execute('SELECT time FROM logs ORDER BY time DESC').fetchone()
last = last[0]

last = datetime.fromtimestamp(last)
now = datetime.now()
if now.day == last.day: exit()

while True:
    rating = input("Day's rating (1-10): ")
    try: rating = int(rating)
    except: continue
    if rating > 10 or rating < 1: check = (input(f"Are you sure it was a(n) {rating}? ").lower() != 'n')
    else: check = True
    if check: break

desc = input("Why was your day a(n) {rating}?\n")

try:
    with conn: c.execute('INSERT INTO logs VALUES (?,?,?)', (int(time()), rating, desc))
except Exception as e: print(e)

