#!/usr/bin/env python3

import random
import sys


def main():
    items = []

    item_no = 1
    start = 1
    while True:
        item = ""
        try:
            if sys.stdin.isatty():
                item = input(f"Item {item_no}: ").strip()
            else:
                item = input().strip()
        except EOFError:
            pass
        if item == "":
            break
        parts, count = item.split("|", 1), 1
        if len(parts) == 2:
            try:
                count = int(parts[1].strip())
                item = parts[0].strip()
            except Exception:
                print("Invalid integer:", parts[1])
                exit(1)
        items += [(item, start, start + count - 1)]
        item_no += 1
        start += count

    if len(items) == 0:
        exit(0)

    N = start - 1
    for i in range(1_000):
        random.randint(1, N)

    if sys.stdin.isatty():
        print()

    item_no = random.randint(1, N)
    for i, (item, start, end) in enumerate(items):
        if start <= item_no <= end:
            print(f"RESULT: Item {i + 1}: {item}")
            break


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print()
        pass
