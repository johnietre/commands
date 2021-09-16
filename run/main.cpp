/*
 * Add flag that changes time unit (sec, ms, ns, etc)
 * Possibly return system return status from command lambda
 * Allow multiple files to be passed
 * Implement -r flag
 * Go through and make sure flags that shouldn't go together can't be put
 * together
 * Add -file-type flag
 * Check exec_args in "go run" command
 * Dissallow flags that can't be used with certain languages
 * Finish parseIncludes
 * Make the way the go and GNU compilers output in the same way
   * Ex: run test/test.go puts the output file in the current directory
     but run test/test.cpp puts it in the directory with the file
 * Print valid filetypes (in help function) using a map iterator
 * Accept just running commands using a flag
   * Ex: run -e {command}
     * Everything after the command is command line arguments for the command
 * Allow bash commands to be input and ran (and timed)
 * Add wasm compilation support
   * Ex: emcc, GOOS=js GOARCH=wasm go build -o ...
 * Update flags (ex. use "--" for non-single letter options)
 * Add option to add go env options (like GOOS=js and GO111MODULE=off)
 * Make temp object files in /temp folder
 */
#include <chrono>
#include <csignal>
#include <cstdlib> // exit
#include <fstream>
#include <iostream> // cout, cerr
#include <map>      // map
#include <string>   // string
using namespace std;

typedef chrono::high_resolution_clock HRC;

enum EXT { C, CPP, F90, HS, GO, JAV, JS, PL, PY, R, RS, SWIFT };

bool compile;
bool del;
bool add_includes;
bool all_hs;
bool no_time;
bool wasm;
EXT file_type;
string filename;
string file_ext;
string exec_name;
string comp_args;
string exec_args;

map<string, EXT> exts = {
    {".c", EXT::C},     {".cc", EXT::CPP},      {".cpp", EXT::CPP},
    {".f", EXT::F90},   {".f75", EXT::F90},     {".f90", EXT::F90},
    {".f95", EXT::F90}, {".go", EXT::GO},       {".hs", EXT::HS},
    {".jav", EXT::JAV}, {".java", EXT::JAV},    {".js", EXT::JS},
    {".pl", EXT::PL},   {".py", EXT::PY},       {".r", EXT::R},
    {".rs", EXT::RS},   {".swift", EXT::SWIFT},
};

void printHelp(int exit_code = 0);
void parseIncludes();
int command(string cmd, bool comp = false);

int main(int argc, char *argv[]) {
  if (argc == 1)
    printHelp();
  string filename = argv[1];
  if (filename == "-b") {
    string cmd = "";
    for (int i = 2; i < argc; i++)
      cmd += string(argv[i]) + " ";
    return command(cmd, false);
  }
  // Check to make sure the filename is valid
  if (filename == "-h") {
    printHelp();
  } else if (filename == ".") {
    file_type = EXT::GO;
    file_ext = ".go";
  } else {
    file_ext = filename.substr(filename.rfind("."));
    if (exts.find(file_ext) != exts.end()) {
      file_type = exts[file_ext];
    } else {
      cerr << "Invalid file type, acceptable types: ";
      for (auto kv : exts)
        cerr << kv.first << " ";
      cerr << '\n';
      return 1;
    }
  }

  // Parse the flags
  string prev = "";
  for (int i = 2; i < argc; i++) {
    string arg = argv[i];
    if (arg == "-c") {
      if (del) {
        cerr << "Cannot use flag \"-c\" with flag \"-d\"\n";
        return 1;
      }
      compile = true;
    } else if (arg == "-d") {
      if (compile) {
        cerr << "Cannot use flag \"-d\" with flag \"-c\"\n";
        return 1;
      }
      del = true;
    } else if (arg == "-hs") {
      if (file_type != EXT::HS) {
        cerr << "\"-hs\" flag must be used with .hs files";
        return 1;
      }
      all_hs = true;
    } else if (arg == "-i") {
      if (file_type != EXT::C && file_type != EXT::CPP) {
        cerr << "Can not use flag \"-i\" with non C/C++ files\n";
        return 1;
      }
      add_includes = true;
    } else if (arg == "-r") {
      cerr << "Not implemented\n";
      printHelp();
    } else if (arg == "-nt") {
      no_time = true;
    } else if (arg == "-wasm") {
      if (del) {
        cerr << "Cannot use flag \"-d\" with flag \"-wasm\"\n";
        return 1;
      }
      wasm = true;
      compile = true;
    } else if (arg == "-b") {
      cerr << "\"-b\" flag must be first argument\n";
      return 1;
    } else {
      if (prev == "-comp-args")
        comp_args = arg;
      else if (prev == "-exec-args")
        exec_args = arg;
      else if (prev == "-o")
        exec_name = arg;
      else if (arg != "-comp-args" && arg != "-exec-args" && arg != "-o") {
        cerr << "Invalid argument: " + arg << '\n';
        printHelp(1);
      }
    }
    prev = arg;
  }

  if (add_includes)
    parseIncludes();
  int status;
  string cmd;
  switch (file_type) {
  case C:
    if (exec_name == "")
      exec_name =
          ((del) ? ".temp" : "") + filename.substr(0, filename.rfind("."));
    comp_args += (comp_args.find("-std=") == string::npos) ? " -std=c18" : "";
    cmd = "gcc -o " + exec_name + " " + filename;
    break;
  case CPP:
    if (exec_name == "")
      exec_name =
          ((del) ? ".temp" : "") + filename.substr(0, filename.rfind("."));
    comp_args +=
        (comp_args.find("-std=") == string::npos) ? " -std=gnu++17" : "";
    cmd = "g++ -o " + exec_name + " " + filename;
    break;
  case F90:
    if (exec_name == "")
      exec_name =
          ((del) ? ".temp" : "") + filename.substr(0, filename.rfind("."));
    cmd = "gfortran -o " + exec_name + " " + filename;
    break;
  case GO:
    if (del)
      return command("go run " + filename + " " + exec_args);
    if (wasm)
      exec_name = (exec_name == "")
                      ? filename.substr(0, filename.rfind('.')) + ".wasm"
                      : exec_name;
    cmd = ((wasm) ? "GOOS=js GOARCH=wasm " : "") + string("go build") +
          ((exec_name == "") ? "" : " -o " + exec_name) + " " + filename;
    if (exec_name == "")
      exec_name = filename.substr(0, filename.rfind("."));
    break;
  case HS:
    if (del)
      return command("runghc " + filename + " " + exec_args);
    cmd = "ghc " + filename;
    exec_name = filename.substr(0, filename.rfind("."));
    break;
  case JAV:
    if (del)
      return command("java " + filename + " " + exec_args);
    status = command("javac " + filename + " " + comp_args, true);
    if (!compile && !status) {
      exec_name = filename.substr(0, filename.rfind("."));
      status = command("java " + exec_name + " " + exec_args);
      if (del)
        remove((exec_name + ".class").c_str());
    }
    return status;
  case JS:
    return command("node " + filename + " " + exec_args);
  case PL:
    return command("perl " + filename + " " + exec_args);
  case PY:
    return command("python3 " + filename + " " + exec_args);
  case R:
    return command("Rscript " + filename + " " + exec_args);
  case RS:
    if (exec_name == "")
      exec_name =
          ((del) ? ".temp" : "") + filename.substr(0, filename.rfind("."));
    cmd = "rustc " + filename + " -o " + exec_name;
    break;
  case SWIFT:
    if (del)
      return command("swift " + filename + " " + exec_args);
    exec_name =
        ((del) ? ".temp" : "") + filename.substr(0, filename.rfind("."));
    cmd = "swiftc " + filename;
    break;
  default:
    printHelp(1);
  }
  // Compile and run the program
  status = command(cmd + " " + comp_args, true);
  if (!all_hs) {
    remove((exec_name + ".o").c_str());
    remove((exec_name + ".hi").c_str());
  }
  if (!compile && !status) {
    status = command("./" + exec_name + " " + exec_args);
    if (del)
      remove(("./" + exec_name).c_str());
  }
  return status;
}

void printHelp(int exit_code) {
  cout << "Usage: run {filename} {flags}\n";
  cout << "Valid file types: .c, .cc, .cpp, .f, .f77, .f90, .f95, .go, .hs, "
          ".jav, .java, .js, .pl, .py, .r, .rs, .swift\n";
  cout << "Flags:\n";
  cout << "-h\tHelp\n";
  cout << "-b {shell code}\tRun code in shell (must all be within a pair of "
          "quotes)\n";
  cout << "-c\tCompile only (all except .py)\n";
  cout << "-d\tDon't leave executable (all except .py)\n";
  cout << "-hs\tLeave all Haskell output files (only works with .hs)\n";
  cout << "-i\tRead file contents to find and compile additional necessary "
          "files\n";
  cout << "-nt\tDo not output compilation and execution time\n";
  cout << "-o {output name}\tThe name of the output executable\n";
  cout << "-r\tRun continuously\n";
  cout << "-wasm\tCompile given file to web assembly\n";
  cout << "-comp-args {compilation args}\tPass arguments to compiler (must all"
          " be put within a pair of quotes)\n";
  cout << "-exec-args {program args}\tPass arguments to program (must all be"
          " put within a pair of quotes)\n";
  exit(exit_code);
}

void parseIncludes() {
  /* TODO */
  return;
}

int command(string cmd, bool comp) {
  if (comp && !no_time)
    cout << "Compiling...\t" << cmd << "\n\n";
  else if (!no_time)
    cout << "Executing...\t" << cmd << "\n\n";
  // Start timing
  HRC::time_point t1 = HRC::now();
  int status = system(cmd.c_str());
  // Stop timing
  HRC::time_point t2 = HRC::now();
  // Calculate the running time
  chrono::duration<double> time_span =
      chrono::duration_cast<chrono::duration<double>>(t2 - t1);
  if (comp) {
    if (no_time)
      cout << '\n';
    else
      cout << "\nCompiliation time: " << time_span.count() << " seconds\n";
    return status;
  } else {
    if (no_time)
      cout << '\n';
    else
      cout << "\nExecution time: " << time_span.count() << " seconds\n";
  }
  return 0;
}
