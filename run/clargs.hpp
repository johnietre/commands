#ifndef CLARGS_HPP
#define CLARGS_HPP

#include <memory>
#include <string>
#include <unordered_map>
#include <vector>
using namespace std;

namespace clargs {

template <typename T>
struct Flag {
  string short_name;
  string long_name;
  string usage;
  T default_value;
  T value;
  bool value_set;
  Flag() = default;
  Flag(string short_name, string long_name, string usage, T default_value) : short_name(short_name), long_name(long_name), usage(usage), default_value(default_value) {}
};

typedef Flag<int> IntFlag;
typedef Flag<double> DoubleFlag;
typedef Flag<bool> BoolFlag;
typedef Flag<string> StringFlag;

typedef unordered_map<string, shared_ptr<IntFlag>> intmap;
typedef unordered_map<string, shared_ptr<DoubleFlag>> doublemap;
typedef unordered_map<string, shared_ptr<BoolFlag>> boolmap;
typedef unordered_map<string, shared_ptr<StringFlag>> stringmap;

class Parser {
private:
  intmap int_flags_;
  doublemap double_flags_;
  boolmap bool_flags_;
  stringmap string_flags_;
  stringmap other_flags_;
  vector<string> other_args_;
public:
  Parser() = default;
  ~Parser() = default;
  Parser(const Parser &rhs) = delete;
  Parser(const Parser &&rhs) = delete;
  Parser& operator= (const Parser &rhs) = delete;
  Parser& operator= (const Parser &&rhs) = delete;

  void parse(int argc, char **argv) {
    string prev;
    for (int i = 0; i < argc; i++) {
      string arg = argv[i];
      if (arg[0] == '-') {
        //
      } else {
        other_args_.push_back(arg);
      }
    }
  }

  void add_int_flag(string short_name, string long_name, string usage, int default_value=0) {
    shared_ptr<IntFlag> flag(new IntFlag(short_name, long_name, usage, default_value));
    int_flags_[short_name] = flag;
    int_flags_[long_name] = flag;
  }

  void add_double_flag(string short_name, string long_name, string usage, double default_value=0.0) {
    shared_ptr<DoubleFlag> flag(new DoubleFlag(short_name, long_name, usage, default_value));
    double_flags_[short_name] = flag;
    double_flags_[long_name] = flag;
  }

  void add_bool_flag(string short_name, string long_name, string usage, bool default_value=false) {
    shared_ptr<BoolFlag> flag(new BoolFlag(short_name, long_name, usage, default_value));
    bool_flags_[short_name] = flag;
    bool_flags_[long_name] = flag;
  }

  void add_string_flag(string short_name, string long_name, string usage, string default_value="") {
    shared_ptr<StringFlag> flag(new StringFlag(short_name, long_name, usage, default_value));
    string_flags_[short_name] = flag;
    string_flags_[long_name] = flag;
  }

  const intmap& int_flags() const {
    return int_flags_;
  }

  const doublemap& double_flags() const {
    return double_flags_;
  }

  const doublemap& bool_flags() const {
    return double_flags_;
  }
  
  const stringmap& string_flags() const {
    return string_flags_;
  }

  const stringmap& other_flags() const {
    return other_flags_;
  }

  const vector<string> other_args() const {
    return other_args_;
  }

  shared_ptr<IntFlag> get_int_flag(string name) {
    return int_flags_[name];
  }

  shared_ptr<DoubleFlag> get_double_flag(string name) {
    return double_flags_[name];
  }

  shared_ptr<BoolFlag> get_bool_flag(string name) {
    return bool_flags_[name];
  }

  shared_ptr<StringFlag> get_string_flag(string name) {
    return string_flags_[name];
  }

  shared_ptr<StringFlag> get_other_flag(string name) {
    return other_flags_[name];
  }

  string get_other_arg(size_t i) {
    return other_args_[i];
  }
};

};

#endif
