#!/usr/bin/env python3
import sqlite3, os, sys
from datetime import datetime
from dateutil import parser
from pathlib import Path

def main():
    path = os.environ.get("JCMDS_PATH")
    if path == "": die('"JCMDS_PATH" environment variable not set')
    path = Path(path) / "dotodo" / "dotodo.db"
    conn = sqlite3.connect(path)
    cur = conn.cursor()
    conn_cur = (conn, cur)

    with conn:
        conn.execute("""CREATE TABLE IF NOT EXISTS items (
            item TEXT,
            time INTEGER,
            completed INTEGER
        );
        """)

    commands = {
        "add": add_command,
        "list": list_command,
        "complete": complete_command,
        "uncomplete": uncomplete_command,
        "delete": delete_command,
        "clear": clear_command,
        "help": help_command,
    }

    args = sys.argv[1:]
    if len(args) == 0: help_command()
    cmd = args[0]
    if cmd not in commands: help_command(args)
    if cmd == "reset": reset_command(path)
    commands[cmd](conn_cur, args[1:])
    conn.close()

def add_command(conn_cur, args):
    if len(args) != 0: help_command(args)
    what = input("Item: ")
    time = input("Time: ")
    try: time = int(parser.parse(time).timestamp())
    except Exception as e: die(e)
    with conn_cur[0]:
        stmt = "INSERT INTO items VALUES (?,?,?)"
        conn_cur[0].execute(stmt, (what, time, 0))

def list_command(conn_cur, args):
    commands = {
        "late": list_late_command,
        "completed": list_completed_command,
        "all": list_all_command,
    }
    if len(args) == 0:
        query = f"SELECT item,time FROM items WHERE completed=0 ORDER BY time"
        items = conn_cur[1].execute(query).fetchall()
        if len(items) == 0: die("no items todo", exit_code=0)
        print_items(items)
        return
    cmd = args[0]
    if cmd not in commands: help_command(args)
    commands[cmd](conn_cur, args[1:])

def list_late_command(conn_cur, args):
    if len(args) != 0: help_command(args)
    now = int(datetime.now().timestamp())
    query = f"SELECT item,time FROM items WHERE completed=0 AND time<{now} ORDER BY time"
    items = conn_cur[1].execute(query).fetchall()
    if len(items) == 0: die("no late items", exit_code=0)
    print_items(items)

def list_completed_command(conn_cur, args):
    if len(args) != 0: help_command(args)
    query = f"SELECT item,time FROM items WHERE completed=1 ORDER BY time"
    items = conn_cur[1].execute(query).fetchall()
    if len(items) == 0: die("no completed items", exit_code=0)
    print_items(items)

def list_all_command(conn_cur, args):
    if len(args) != 0: help_command(args)
    query = f"SELECT item,time,completed FROM items ORDER BY time"
    items = conn_cur[1].execute(query).fetchall()
    if len(items) == 0: die("no items", exit_code=0)
    for (item, time, completed) in items:
        checkmark = chr(0x2705)+" " if completed == 1 else ""
        print(f"{checkmark}{timestring(time)} | {item}")

def complete_command(conn_cur, args):
    if len(args) != 0: help_command(args)
    query = f"SELECT item,time FROM items WHERE completed=0 ORDER BY time"
    items = conn_cur[1].execute(query).fetchall()
    if len(items) == 0: die("no items todo", exit_code=0)
    for (i, (item, time)) in enumerate(items):
        time = datetime.fromtimestamp(time)
        time = time.strftime("%I:%M %p %m/%d/%y")
        print(f"{i+1}) {time} | {item}")
    choice = -1
    try: choice = int(input("Choice (0 to cancel): "))
    except: die("invalid choice")
    if choice == 0: return
    if choice < 1 or choice > len(items): die("invalid choice")
    (item, time) = items[choice-1]
    with conn_cur[0]:
        stmt = f'UPDATE items SET completed=1 WHERE completed=0 AND item="{item}" AND time={time}'
        conn_cur[0].execute(stmt)

def uncomplete_command(conn_cur, args):
    if len(args) != 0: help_command(args)
    query = f"SELECT item,time FROM items WHERE completed=1 ORDER BY time"
    items = conn_cur[1].execute(query).fetchall()
    if len(items) == 0: die("no completed items", exit_code=0)
    for (i, (item, time)) in enumerate(items):
        time = datetime.fromtimestamp(time)
        time = time.strftime("%I:%M %p %m/%d/%y")
        print(f"{i+1}) {time} | {item}")
    choice = -1
    try: choice = int(input("Choice (0 to cancel): "))
    except: die("invalid choice")
    if choice == 0: return
    if choice < 1 or choice > len(items): die("invalid choice")
    (item, time) = items[choice-1]
    with conn_cur[0]:
        stmt = f'UPDATE items SET completed=0 WHERE completed=1 AND item="{item}" AND time={time}'
        conn_cur[0].execute(stmt)

def delete_command(conn_cur, args):
    if len(args) == 0: help_command("missing delete command")
    commands = {
        "completed": delete_completed_command,
        "uncompleted": delete_uncompleted_command,
    }
    cmd = args[0]
    if cmd not in commands: help_command(args)
    commands[cmd](conn_cur, args[1:])

def delete_completed_command(conn_cur, args):
    if len(args) != 0: help_command(args)
    query = f"SELECT item,time FROM items WHERE completed=1 ORDER BY time"
    items = conn_cur[1].execute(query).fetchall()
    if len(items) == 0: die("no completed items", exit_code=0)
    for (i, (item, time)) in enumerate(items):
        time = datetime.fromtimestamp(time)
        time = time.strftime("%I:%M %p %m/%d/%y")
        print(f"{i+1}) {time} | {item}")
    choice = -1
    try: choice = int(input("Choice (0 to cancel): "))
    except: die("invalid choice")
    if choice == 0: return
    if choice < 1 or choice > len(items): die("invalid choice")
    (item, time) = items[choice-1]
    with conn_cur[0]:
        stmt = f'DELETE FROM items WHERE completed=1 AND item="{item}" AND time={time}'
        conn_cur[0].execute(stmt)

def delete_uncompleted_command(conn_cur, args):
    if len(args) != 0: help_command(args)
    query = f"SELECT item,time FROM items WHERE completed=0 ORDER BY time"
    items = conn_cur[1].execute(query).fetchall()
    if len(items) == 0: die("no completed items", exit_code=0)
    for (i, (item, time)) in enumerate(items):
        time = datetime.fromtimestamp(time)
        time = time.strftime("%I:%M %p %m/%d/%y")
        print(f"{i+1}) {time} | {item}")
    choice = -1
    try: choice = int(input("Choice (0 to cancel): "))
    except: die("invalid choice")
    if choice == 0: return
    if choice < 1 or choice > len(items): die("invalid choice")
    (item, time) = items[choice-1]
    with conn_cur[0]:
        stmt = f'DELETE FROM items WHERE completed=0 AND item="{item}" AND time={time}'
        conn_cur[0].execute(stmt)

def clear_command(conn_cur, args):
    if len(args) == 0: help_command("missing clear command")
    commands = {
        "completed": clear_completed_command,
        "uncompleted": clear_uncompleted_command,
        "all": clear_all_command,
    }
    cmd = args[0]
    if cmd not in commands: help_command(args)
    commands[cmd](conn_cur, args[1:])

def clear_completed_command(conn_cur, args):
    if len(args) != 0: help_command(args)
    choice = input("Clear completed items? [Y/n] ").lower()
    if choice != "y": return
    with conn_cur[0]:
        stmt = "DELETE FROM items WHERE completed=1"
        conn_cur[0].execute(stmt)

def clear_uncompleted_command(conn_cur, args):
    if len(args) != 0: help_command(args)
    choice = input("Clear uncompleted items? [Y/n] ").lower()
    if choice != "y": return
    with conn_cur[0]:
        stmt = "DELETE FROM items WHERE completed=0"
        conn_cur[0].execute(stmt)

def clear_all_command(conn_cur, args):
    if len(args) != 0: help_command(args)
    choice = input("Clear all items? [Y/n] ").lower()
    if choice != "y": return
    with conn_cur[0]:
        stmt = "DELETE FROM items"
        conn_cur[0].execute(stmt)

def reset_command(path):
    choice = input("Reset program? [Y/n] ").lower()
    if choice != 'y': return
    os.remove(path)

def help_command(*args):
    if len(args) != 0:
        arg = args[0]
        if type(arg) == str:
            print(args)
        elif type(arg) == list:
            print(f"unknown command(s): {' '.join(arg)}")
    print("Usage: dotodo <command> [subcommands]")
    print("    add\t\t\tAdds an task")
    print("    list\t\tLists all uncompleted tasks")
    print("\tlate\t\tLists all past-due uncompleted tasks")
    print("\tcompleted\tLists all completed tasks")
    print("\tall\t\tLists all tasks")
    print("    delete")
    print("\tcompleted\tDeletes a completed task")
    print("\tuncompleted\tDeletes a uncompleted task")
    print("    clear")
    print("\tcompleted\tClears all completed tasks")
    print("\tuncompleted\tClears all uncompleted tasks")
    print("\tall\t\tClears all tasks")
    print("    help\t\tPrints help screen")
    print("    reset\tResets the program")
    sys.exit(1)

def print_items(items):
    for (item, time) in items: print(f"{timestring(time)} | {item}")

def timestring(timestamp):
    return datetime.fromtimestamp(timestamp).strftime("%I:%M %p %a, %b %d, %Y")

def die(*args, exit_code=1, **kwargs):
    print(*args, file=sys.stderr, **kwargs)
    sys.exit(exit_code)

if __name__ == "__main__": main()
