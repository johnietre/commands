#![allow(dead_code)]
use std::process::Command as PCommand;
use std::time::Instant;

mod clargs;
use clargs::*;

fn main() {
}

// A list of the accepted file types
enum FileType {
    C,
    CPP,
    F90,
    GO,
    HS,
    JAV,
    JS,
    PL,
    PY,
    R,
    RS,
    SWIFT
}

fn get_file_type(file_name: String) {
    //match file_name.extension() {
    match file_name.as_str() {
        "c" => (),
        "cc" | "cxx" | "cpp" => (),
        "f" | "f75" | "f90" | "f95" => (),
        "go" => (),
        "hs" => (),
        "jav" | "java" => (),
        "js" => (),
        "pl" => (),
        "py" => (),
        "r" => (),
        "rs" => (),
        "swift" => (),
        _ => (),
    }
}

enum TimeType {
    Nano,
    Micro,
    Milli,
    Sec,
}

struct Command {
    file_name: String,
    file_type: Option<FileType>,
    exec_name: String,
    comp_args: Vec<String>,
    exec_args: Vec<String>,

    delete: bool,
    time_type: TimeType,
    bash_only: bool,
    compile: bool,
    temp: bool,
    all_hs: bool,
    no_time: bool,
    wasm: bool,
}

impl Command {
    fn new() {}

    fn run(&mut self) {
        use FileType::*;
        match self.file_type {
            C => {
                if self.exec_name == "" {
                    if self.delete {
                        self.exec_name = ;
                    }
                }
            },
            CPP =>,
            F90 =>,
            GO =>,
            HS =>,
            JAV =>,
            JS =>,
            PL =>,
            PY =>,
            R =>,
            RS =>,
            SWIFT =>,
        }
    }

    fn run_c(&self) {
        //
    }

    fn run_cpp(&self) {}

    fn run_f90(&self) {}

    fn run_go(&self) {}

    fn run_hs(&self) {}

    fn run_jav(&self) {}

    fn run_js(&self) {}

    fn run_pl(&self) {}

    fn run_py(&self) {}

    fn run_r(&self) {}

    fn run_rs(&self) {}

    fn run_swift(&self) {}

    fn exec(&self, cmd: &str, compile: bool) -> i32 {
        use TimeType::*;

        if !self.no_time {
            println!("{}...\t{}\n", if compile { "Compiling" } else { "Executing" }, cmd);
        }
        // Create the command
        let mut cmds = cmd.split_whitespace();
        let mut command = PCommand::new(cmds.next().unwrap());
        command.args(cmds);
        // Run and time the command
        let start = Instant::now();
        let status = command.status().unwrap().code().unwrap();
        let dur = start.elapsed();
        let time_str = match self.time_type {
            Sec => format!("{} seconds", dur.as_secs_f64()),
            Milli => format!("{} milliseconds", dur.as_millis()),
            Micro => format!("{} microseconds", dur.as_micros()),
            Nano => format!("{} nanoseconds", dur.as_nanos()),
        };
        if !self.no_time {
            println!("{} time: {}", if compile { "Execution" } else { "Compilation" }, time_str);
        }
        status
    }
}
