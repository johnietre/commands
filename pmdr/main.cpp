// Display "Paused" in place of "Work!" or "Rest!" when paused
// Allow "s" to be pressed to skip current mode (skip work to rest or vice versa)
//#include "ansi_codes.h"
#define MINIAUDIO_IMPLEMENTATION
#include "miniaudio.h"
#include "terminal.h"
#include <atomic>
#include <chrono>
#include <cstdlib>
#include <csignal>
#include <iostream>
#include <stdexcept>
#include <string>
#include <thread>
using namespace std;

typedef unsigned int ui;

atomic<bool> working(true);
atomic<bool> paused(false);
atomic<bool> skip(false);
ma_engine engine;
ma_sound work_done_sound, rest_done_sound;

void pt(ui secs, ui prev);
void beep();
void listenPause();
void sound();

void atexit_handler() {
  reset_terminal_mode();
}

int main(int argc, char **argv) {
  // Set the work and rest minutes (get from command-line)
  ui workMin = 25, restMin = 5;
  for (int i = 1; i < argc; i++) {
    string arg = argv[i];
    if (arg == "-h" || arg == "--help") {
      printf("Usage: pmdr [options]\n");
      printf("Options:\n");
      printf("    -w work-minutes\t\tNumber of minutes in the work period (default is 25)\n");
      printf("    -r rest-minutes\t\tNumber of minutes in the rest period (default is 5)\n");
      return 0;
    } else if (arg == "-w") {
      if (++i == argc) {
        printf("Must provide number of minutes\n");
        return 0;
      }
      try {
        workMin = stoul(argv[i]);
      } catch (const exception &e) {
        printf("Invalid minutes input\n");
        return 1;
      }
    } else if (arg == "-r") {
      if (++i == argc) {
        printf("Must provide number of minutes\n");
        return 1;
      }
      try {
        restMin = stoul(argv[i]);
      } catch (const exception &e) {
        printf("Invalid minutes input\n");
        return 1;
      }
    } else {
      printf("Invalid flag: %s\n", argv[i]);
      return 1;
    }
  }

  // Initialize the sounds
  ma_result result;
  if ((result = ma_engine_init(NULL, &engine)) != MA_SUCCESS)
    return result;

  // Set the terminal mode to raw
  set_terminal_mode();
  if (atexit(atexit_handler) != 0)
    cerr << "Error setting atexit handler to reset terminal\n";
  // Start the threads for the beeper and pauser
  thread soundPlayer(sound);
  thread pauser(listenPause);
  // Print the controls
  puts("p, enter = pause; q, ctrl-c = quit; s = skip");
  // Start the timer
  int period = 1;
  printf("Work! (period %d)\n", 1);
  while (1) {
    for (ui t = 60 * ((working) ? workMin : restMin), prev = 0; t > 0 && !skip; prev = t, t--) {
      while (paused);
      pt(t, prev);
      this_thread::sleep_for(std::chrono::seconds(1));
    }
    working = !working;
    if (working) {
      period++;
    }
    printf(CUR_UP);
    printf(CUR_LEFTN(40));
    printf("%s (period %d)", (working) ? "Work!" : "Rest!", period);
    printf(CUR_DOWN);
    printf(CUR_LEFTN(40));
    skip = false;
  }
  return 0;
}

void pt(ui secs, ui prev) { // prints the time
  ui diff = (prev / 10) - (secs / 10);
  if (diff) {
    if ((prev / 60) - (secs / 60)) {
      printf(CUR_LEFTN(5));
      printf("%02u:", secs / 60);
    } else {
      printf(CUR_LEFTN(2));
    }
    printf("%02u", secs % 60);
  } else {
    printf(CUR_LEFT);
    printf("%u", secs % 10);
  }
  fflush(stdout);
}

void sound() {
  bool old = true;
  while (1) {
    while (working == old);
    //
    old = working;
  }
}

void listenPause() {
  while (1) {
    switch (getchar()) {
      case '\n':
      case 'p':
        paused = !paused;
        break;
      case 'q':
        puts("");
        reset_terminal_mode();
        exit(0);
      case 's':
        skip = true;
    }
  }
}
