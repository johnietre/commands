#!/usr/bin/env python3

import json, openai, os, sys
from pathlib import Path

source_dir = None

def main():
    global source_dir

    openai.organization = os.getenv("OPENAI_API_ORG")
    openai.api_key = os.getenv("OPENAI_API_KEY")

    file_path = Path(__file__)
    source_dir = os.path.join(
        os.path.dirname(file_path),
        "..", "aime", 
    )
    chat_completion()

def chat_completion():
    global source_dir

    chat_comp_path = os.path.join(source_dir, "chats_comp_msgs.json")
    resp_output_path = os.path.join(source_dir, "responses.json")

    msgs = []
    try:
        with open(chat_comp_path, "r") as f: msgs = json.load(f)
    except FileNotFoundError: pass
    except Exception as e:
        print("Error opening or reading json file:", e)
        exit(1)

    if len(msgs) == 0:
        msgs.append({"role": "system", "content": "You are an AI assistant."})

    send_history = True
    if len(sys.argv) > 1:
        send_history = sys.argv[1].lower() in ["false", "f"]
    while True:
        print("What's your message (triple enter to send, blank message to exit)?")
        new_msg = {"role": "user", "content": ""}
        blank = False
        while True:
            line = input()
            if len(line) == 0:
                if blank: break
                blank = True
            else:
                blank = False
                new_msg["content"] += line + "\n"
        if len(new_msg["content"].strip()) == 0: exit(0)
        msgs.append(new_msg)

        print("Sending...")
        print()
        try:
            response = openai.ChatCompletion.create(
                model="gpt-3.5-turbo",
                messages=msgs if send_history else [],
            )
        except Exception as e:
            print("Error sending:", e)
            exit(1)

        with open(resp_output_path, "a+") as resp_file:
            resp_file.write(json.dumps(response) + "\n")

        resp_msg = response.choices[0].message
        msgs.append(resp_msg)
        with open(chat_comp_path, "w+") as msgs_file:
            msgs_file.write(json.dumps(msgs))
        print(resp_msg["content"])
        print()

if __name__ == "__main__": main()
