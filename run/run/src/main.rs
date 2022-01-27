#![allow(dead_code)]
use std::path::PathBuf;
use std::process::{exit, Command};
use std::time::Instant;

#[macro_use]
extern crate lazy_static;

lazy_static! {
    static ref temp_dir: PathBuf = std::env::temp_dir();
}

fn main() {
    let mut cmd = Cmd::new();
    cmd.exec("python3 test.py", false);
}

// A list of the accepted file types
#[derive(Clone, Copy)]
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

impl FileType {
    fn file_type_from_string(s: String) -> Option<Self> {
        use FileType::*;
        match s.as_str() {
            "C" => Some(C),
            "CPP" => Some(CPP),
            "F90" => Some(F90),
            "GO" => Some(GO),
            "HS" => Some(HS),
            "JAV" => Some(JAV),
            "JS" => Some(JS),
            "PL" => Some(PL),
            "PY" => Some(PY),
            "R" => Some(R),
            "RS" => Some(RS),
            "SWIFT" => Some(SWIFT),
            _ => None,
        }
    }

    fn file_type_from_ext(file_name: String) -> Option<Self> {
        use FileType::*;
        match file_name.as_str() {
            "c" => Some(C),
            "cc" | "cxx" | "cpp" => Some(CPP),
            "f" | "f75" | "f90" | "f95" => Some(F90),
            "go" => Some(GO),
            "hs" => Some(HS),
            "jav" | "java" => Some(JAV),
            "js" => Some(JS),
            "pl" => Some(PL),
            "py" => Some(PY),
            "r" => Some(R),
            "rs" => Some(RS),
            "swift" => Some(SWIFT),
            _ => None,
        }
    }
}

enum TimeType {
    Nano,
    Micro,
    Milli,
    Sec,
}

impl Default for TimeType {
    fn default() -> Self {
        Self::Sec
    }
}

#[derive(Default)]
struct Cmd {
    file_name: String,
    file_type: Option<FileType>,
    exec_name: String,
    comp_args_before: String,
    comp_args_after: String,
    exec_args: String,

    delete: bool,
    time_type: TimeType,
    bash_only: bool,
    compile_only: bool,
    temporary: bool,
    all_haskell_out_files: bool,
    no_time: bool,
    wasm: bool,
}

impl Cmd {
    fn new() -> Self {
        Self {
            ..Default::default()
        }
    }

    fn run(&mut self) {
        use FileType::*;
        match self.file_type.unwrap() {
            C => self.run_c(),
            CPP => self.run_c(),
            F90 => self.run_f90(),
            GO => self.run_go(),
            HS => self.run_hs(),
            JAV => self.run_jav(),
            JS => self.run_js(),
            PL => self.run_pl(),
            PY => self.run_py(),
            R => self.run_r(),
            RS => self.run_rs(),
            SWIFT => self.run_swift(),
            _ => unreachable!(),
        }
    }

    fn run_c(&self) {
        //
    }

    fn run_cpp(&self) {
        //
    }

    fn run_f90(&self) {
        //
    }

    fn run_go(&self) {
    }

    fn run_hs(&self) {
        if self.delete {
            // Run runghc if no object file is desired
            exit(
                self.exec(
                    &format!("runghc {} {}", self.file_name, self.exec_args),
                    false),
            );
            return;
        }
        // Build and run the compilation commmand
        let cmd = format!(
            "ghc -o {} {} {} {}",
            self.exec_name,
            self.comp_args_before,
            self.file_name,
            self.comp_args_after,
        );
        let status = self.exec(&cmd, true);
        if self.compile_only || status != 0 {
            exit(status);
        }
        // Build and run the execution command
        let cmd = if self.exec_name.contains("/") {
            self.exec_name
        } else {
            "./".to_owned() + &self.exec_name
        } + &self.exec_args;
        exit(self.exec(&cmd, false));
    }

    fn run_jav(&self) {
    }

    fn run_js(&self) {
        exit(self.exec(
            &format!("node {} {}", self.file_name, self.exec_args),
            false));
    }

    fn run_pl(&self) {
        self.exec(
            &format!("perl {} {}", self.file_name, self.exec_args),
            false);
    }

    fn run_py(&self) {
        self.exec(
            &format!("python3 {} {}", self.file_name, self.exec_args),
            false);
    }

    fn run_r(&self) {
        self.exec(
            &format!("Rscript {} {}", self.file_name, self.exec_args),
            false);
    }

    fn run_rs(&self) {
        //
    }

    fn run_swift(&self) {
        unimplemented!();
    }

    fn exec(&self, cmd: &str, compile: bool) -> i32 {
        use TimeType::*;

        if !self.no_time {
            println!("{}...\t{}\n", if compile { "Compiling" } else { "Executing" }, cmd);
        }
        // Create the command
        let mut cmds = cmd.split_whitespace();
        let mut command = Command::new(cmds.next().unwrap());
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
            println!("\n{} time: {}", if compile { "Execution" } else { "Compilation" }, time_str);
        }
        status
    }
}
