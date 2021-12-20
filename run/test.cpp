#include "clargs.hpp"
#include <iostream>
#include <string>
using namespace std;

int main(int argc, char **argv) {
  clargs::Parser parser;
  parser.parse(argc, argv);
  return 0;
}
