#!/usr/bin/env python3
# TODO: Implement unimplemented items (from help printout)
# TODO: Allow item and time to be passed through the command line program call
import sqlite3, os, sys
from datetime import datetime, timedelta
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
    if cmd == "reset": reset_command(path)
    elif cmd == "version": version_command()
    elif cmd not in commands: help_command(args)
    else: commands[cmd](conn_cur, args[1:])
    conn.close()

def add_command(conn_cur, args):
    if len(args) > 2: help_command(args)
    days = [
            "mon", "tue", "wed", "thu", "fri", "sat", "sun",
            "monday", "tuesday", "wednesday", "thursday", "friday", "saturday",
            "sunday"
            ]
    what, time = "", ""
    # Get the what (item) and time
    if len(args) > 0:
        what = args[0]
        if len(args) == 2: time = args[1]
        else: time = input("Time: ").lower()
    else:
        what = input("Item: ")
        time = input("Time: ").lower()
    hour_min = time.split(" ", 1)
    hour, minu, sec = 23, 59, 59
    if len(hour_min) != 1:
        times_tup = None
        if hour_min[0] != "" and ":" in hour_min[0]:
            times_tup = parse_hour_min(hour_min[0])
            time = hour_min[1]
        else:
            times_tup = parse_hour_min(hour_min[1])
            time = hour_min[0]
        if times_tup != None: hour, minu, sec = times_tup
    else: hour_min = None
    if time == "today":
        today = datetime.now().date()
        time = datetime(today.year, today.month, today.day, hour, minu, sec)
    elif time == "tomorrow":
        tomorrow = datetime.now().date() + timedelta(1)
        time = datetime(tomorrow.year, tomorrow.month, tomorrow.day, hour, minu, sec)
    elif time in days:
        weekday = days.index(time) % 7
        today = datetime.now().date()
        diff = weekday - today.weekday()
        new_day = None
        if diff < 0: today + timedelta(7+diff)
        elif diff == 0: new_day = today + timedelta(7)
        else: new_day = today + timedelta(diff)
        time = datetime(new_day.year, new_day.month, new_day.day, hour, minu, sec)
    else:
        try: time = parser.parse(time)
        except Exception as e: die(e)
    with conn_cur[0]:
        stmt = "INSERT INTO items VALUES (?,?,?)"
        conn_cur[0].execute(stmt, (what, int(time.timestamp()), 0))

def list_command(conn_cur, args):
    commands = {
        "late": list_late_command,
        "completed": list_completed_command,
        "all": list_all_command,
        "for": None,
        "before": None,
        "after": None,
        "between": None,
    }
    if len(args) == 0:
        now = datetime.now()
        query = f"SELECT item,time FROM items WHERE completed=0 ORDER BY time"
        items = conn_cur[1].execute(query).fetchall()
        if len(items) == 0: die("no items todo", exit_code=0)
        for (item, time) in items:
            x = "" if datetime.fromtimestamp(time) > now else chr(0x274C) + " "
            print(f"{x}{timestring(time)} | {item}")
        return
    cmd = args[0]
    if cmd not in commands:
        # TODO
        if cmd == "for":
            list_for()
        else:
            list_between()
        help_command(args)
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
    now = datetime.now()
    query = f"SELECT item,time,completed FROM items ORDER BY time"
    items = conn_cur[1].execute(query).fetchall()
    if len(items) == 0: die("no items", exit_code=0)
    for (item, time, completed) in items:
        mark = ""
        if completed == 1: mark = chr(0x2705) + " "
        elif datetime.fromtimestamp(time) <= now: mark = chr(0x274C) + " "
        print(f"{mark}{timestring(time)} | {item}")

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

def version_command():
    print("versionless")

def help_command(*args):
    if len(args) != 0:
        arg = args[0]
        if type(arg) == str:
            print(args)
        elif type(arg) == list:
            print(f"unknown command(s): {' '.join(arg)}")
    print("Usage: dotodo <command> [subcommands]")
    print('    add [item [time]]\tAdds a task ("today" or "tomorrow" can be given as the time)')
    print("    list\t\tLists all uncompleted tasks")
    print("\tlate\t\tLists all past-due uncompleted tasks")
    print("\tcompleted\tLists all completed tasks")
    print("\tall\t\tLists all tasks")
    print("    delete")
    print("\tcompleted\tDeletes a completed task")
    print("\tuncompleted\tDeletes an uncompleted task")
    print("    clear")
    print("\tcompleted\tClears all completed tasks")
    print("\tuncompleted\tClears all uncompleted tasks")
    print("\tall\t\tClears all tasks")
    print("    reset\t\tResets the program")
    print("    version\t\tPrints dotodo version")
    print("    help\t\tPrints help screen")
    print()
    print('Commands can be applied for a specific date by adding "for [date]" after it')
    print('Commands can be applied to before a specific date/time by adding "before [date/time]" after it')
    print('Commands can be applied to after a specific date/time by adding "after [date/time]" after it')
    print('Commands can be applied to between specific dates/times by adding "between [start date/time] [end date/time]" after it')
    print("These have yet to be implemented!")
    sys.exit(1)

def print_items(items):
    for (item, time) in items: print(f"{timestring(time)} | {item}")

def timestring(timestamp):
    return datetime.fromtimestamp(timestamp).strftime("%I:%M %p %a, %b %d, %Y")

def parse_hour_min(time):
    try:
        time = parser.parse(time)
        return (time.hour, time.minute, time.second)
    except: return None

def die(*args, exit_code=1, **kwargs):
    print(*args, file=sys.stderr, **kwargs)
    sys.exit(exit_code)

if __name__ == "__main__": main()
