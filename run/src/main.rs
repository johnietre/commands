// TODO: Default flags passed
// TODO: Clean and refactor
// TODO: Delete output after --no-out flag
// TODO: Stuff like segmentation fault message not being output
// Building go directory not working
// In order to pass flags through comp or exec flag, use: --[comp/exec]_arg=--flag
// To pass multiple arguments in one call, use: --[comp/exec]_arg={--flag1,arg1,-opt}
// Add help

use clap::Parser;
use std::collections::hash_map::DefaultHasher;
use std::env::{self, temp_dir};
use std::fs::canonicalize;
use std::hash::{Hash, Hasher};
use std::path::PathBuf;
use std::process::exit;
use std::time::Instant;

pub mod ansi;

pub mod app;

pub mod args;
use args::*;

pub mod deferrer;

pub mod file_type;

pub mod run_funcs;
use run_funcs::{run_executable, run_multiple};

pub mod watcher;

#[macro_export]
macro_rules! die {
    ($code:expr, $($args:tt)*) => ({
        eprintln!($($args)*);
        ::std::process::exit($code)
    });
}

fn main() {
    // Check to see if the first/second argument is the bash flag and, if so, pass the remaining
    // arguments to bash.
    let first2v = env::args().skip(1).take(2).collect::<Vec<_>>();
    let first2 = first2v.iter().map(|s| s.as_str()).collect::<Vec<_>>();
    let mut args = match first2.as_slice() {
        ["--no-time", "-b" | "--bash"] | ["-b" | "--bash", _] => {
            let no_time = first2[0] == "--no-time";
            let args = Args {
                file_names: vec!["-c".into()],
                exec_args: env::args().skip(if no_time { 3 } else { 2 }).collect(),
                file_type: Some(file_type::FileType::BASH),
                no_time,
                bash: true,
                ..Default::default()
            };
            args
        }
        _ => {
            let args = Args::parse();
            if args.bash {
                die!(1, "bash flag must be the first argument");
            }
            args
        }
    };

    // Get the file type
    let file_type = if let Some(t) = args.file_type {
        if args.file_names.len() == 0 && !t.can_have_none() {
            exit(run_multiple(&args))
        }
        t
    } else {
        let file_name = args
            .file_names
            .get(0)
            .map_or_else(|| exit(run_multiple(&args)), |name| name);
        file_name
            .extension()
            .unwrap_or_default()
            .to_string_lossy()
            .parse()
            .unwrap_or_else(|err| die!(1, "error parsing file type: {}", err))
    };
    // Check the file name(s)
    if args.file_names.len() > 1 && !file_type.can_have_multiple() {
        die!(1, "only one file allowed")
    }
    args.file_type.replace(file_type);
    if args.no_out {
        //args.temp = true;
    }
    // Get the temp output name
    if args.temp {
        let mut hasher = DefaultHasher::new();
        if let Some(name) = args.file_names.get(0) {
            if let Ok(path) = canonicalize(name) {
                path.hash(&mut hasher);
            }
        }
        Instant::now().hash(&mut hasher);
        let name = format!("{:x}", hasher.finish());
        args.output_name.replace(temp_dir().join(name));
    } else if args.output_name.is_none() && args.file_names.len() != 0 {
        if let Some(mut name) = args.file_names[0].file_stem().map(PathBuf::from) {
            if cfg!(windows) {
                // TODO: Check to see if extension already set
                name.set_extension(".exe");
            }
            args.output_name.replace(name);
        }
    }
    if args.wasm {
        args.compile_only = true;
    }
    // Get the func to run the file
    let run_func = file_type.run_func();
    if args.watch && args.compile_only {
        // TODO: Exit code
        die!(2, "cannot pass a compile_only flag and watch flag");
    }
    // Run the compilation or execution
    // If the file was compiled and needs to be run, run the executable,
    // otherwise, exit
    if let Some(status) = run_func(&args) {
        exit(status)
    } else if !args.compile_only {
        exit(run_executable(&args))
    }
}
