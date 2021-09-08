#ifndef FLAGS_HPP
#define FLAGS_HPP

#include <map>
#include <string>
using namespace std;

static int var = 0;

namespace flags{

class FlagSet {
private:
  enum _flag_type {
    INT,
    UNSIGNED_INT,
    LONG,
    UNSIGNED_LONG,
    FLOAT,
    DOUBLE,
    BOOL,
    CHAR,
    STRING,
  };
  struct _flag_info {
    _flag_type type;
    void *ptr;
  };
  map<string, _flag_info> _flags;
  int _argc;
  char **_argv;
public:
  FlagSet(int argc, char **argv) {
    _argc = argc;
    _argv = argv;
  }
  bool set_flag(string flag_name, int *ptr) {
    _flags[flag_name] = {INT, static_cast<void*>(ptr)};
  }
  bool set_flag(string flag_name, unsigned int *ptr) {
    _flags[flag_name] = {UNSIGNED_INT, static_cast<void*>(ptr)};
  }
  bool set_flag(string flag_name, long *ptr) {
    _flags[flag_name] = {LONG, static_cast<void*>(ptr)};
  }
  bool set_flag(string flag_name, unsigned long *ptr) {
    _flags[flag_name] = {UNSIGNED_LONG, static_cast<void*>(ptr)};
  }
  bool set_flag(string flag_name, float *ptr) {
    _flags[flag_name] = {FLOAT, static_cast<void*>(ptr)};
  }
  bool set_flag(string flag_name, double *ptr) {
    _flags[flag_name] = {DOUBLE, static_cast<void*>(ptr)};
  }
  bool set_flag(string flag_name, bool *ptr) {
    _flags[flag_name] = {BOOL, static_cast<void*>(ptr)};
  }
  bool set_flag(string flag_name, char *ptr) {
    _flags[flag_name] = {CHAR, static_cast<void*>(ptr)};
  }
  bool set_flag(string flag_name, string *ptr) {
    _flags[flag_name] = {STRING, static_cast<void*>(ptr)};
  }
  void parse() {
    for (int i = 1; i < _argc; i++) {
      //
    }
  }
};

};
#endif
