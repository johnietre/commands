use crate::args::*;
use crate::die;
use crate::file_type::*;
use std::fs::remove_file;
use std::process::{Child, Command};
use std::time::Instant;

pub type RunFunc = fn(&Args) -> Option<i32>;

pub fn run_bash(args: &Args) -> Option<i32> {
    let mut cmd = Command::new("bash");
    cmd.arg(&args.file_names[0]).arg(&args.exec_args.join(" "));
    Some(execute(&mut cmd, args, false))
}

pub fn run_c(args: &Args) -> Option<i32> {
    let mut cmd = new_compile_cmd(args.prog("gcc"), args);
    if args.default_flags {
        cmd.args(&["-std=c18"]);
    }
    if args.watch {
        return Some(run_watch(args, new_exec_cmd(args), Some(cmd), &[]));
    }
    let status = execute(&mut cmd, args, true);
    if status == 0 {
        None
    } else {
        Some(status)
    }
}

pub fn run_cpp(args: &Args) -> Option<i32> {
    let mut cmd = new_compile_cmd(args.prog("g++"), args);
    if args.default_flags {
        cmd.args(&["-std=gnu++17"]);
    }
    if args.watch {
        return Some(run_watch(args, new_exec_cmd(args), Some(cmd), &[]));
    }
    let status = execute(&mut cmd, args, true);
    if status == 0 {
        None
    } else {
        Some(status)
    }
}

pub fn run_f90(args: &Args) -> Option<i32> {
    let mut cmd = new_compile_cmd(args.prog("gfortran"), args);
    if args.watch {
        return Some(run_watch(args, new_exec_cmd(args), Some(cmd), &[]));
    }
    let status = execute(&mut cmd, args, true);
    if status == 0 {
        None
    } else {
        Some(status)
    }
}

pub fn run_go(args: &Args) -> Option<i32> {
    if args.no_out {
        // TODO: Make this line active and one below not
        //if args.no_out && !args.temp {
        let mut cmd = Command::new("go");
        cmd.arg("run")
            .args(&args.comp_args)
            .args(&args.file_names)
            .args(&args.exec_args);
        if args.watch {
            return Some(run_watch(args, cmd, None, &[]));
        }
        return Some(execute(&mut cmd, args, false));
    }
    // TODO: If compiling directory, get name of directory
    let mut cmd = Command::new("go");
    cmd.arg("build")
        .arg("-o")
        .arg(
            args.output_name
                .as_ref()
                .unwrap_or_else(|| die!(2, "output name not set")),
        )
        .args(&args.comp_args)
        .args(&args.file_names);
    if args.wasm {
        cmd.envs([("GOOS", "js"), ("GOARCH", "wasm")]);
    }
    if args.watch {
        return Some(run_watch(args, new_exec_cmd(args), Some(cmd), &[]));
    }
    let status = execute(&mut cmd, args, true);
    if status == 0 {
        None
    } else {
        Some(status)
    }
}

pub fn run_hs(args: &Args) -> Option<i32> {
    if args.no_out {
        let mut cmd = Command::new(args.prog("runghc"));
        cmd.arg(&args.file_names[0]).args(&args.exec_args);
        if args.watch {
            return Some(run_watch(args, cmd, None, &[]));
        }
        return Some(execute(&mut cmd, args, false));
    }
    let prog = args.prog("ghc");
    let exts: &[&str] = if !args.keep_all_out && prog == "ghc" {
        &["o", "hi"]
    } else {
        &[]
    };
    let mut cmd = new_compile_cmd(&prog, args);
    if args.watch {
        return Some(run_watch(args, new_exec_cmd(args), Some(cmd), exts));
    }
    let status = execute(&mut cmd, args, true);
    if exts.len() != 0 {
        delete_files(exts, args);
    }
    if status == 0 {
        None
    } else {
        Some(status)
    }
}

pub fn run_jav(args: &Args) -> Option<i32> {
    // TODO: Properly format javac arguments
    let mut cmd = Command::new("javac");
    cmd.args(&args.file_names).args(&args.comp_args);
    if args.watch {
        return Some(run_watch(args, new_exec_cmd(args), Some(cmd), &[]));
    }
    let status = execute(&mut cmd, args, true);
    if status == 0 {
        None
    } else {
        Some(status)
    }
}

pub fn run_jl(args: &Args) -> Option<i32> {
    let mut cmd = Command::new(args.prog("julia"));
    cmd.arg(&args.file_names[0]).args(&args.exec_args);
    if args.watch {
        return Some(run_watch(args, cmd, None, &[]));
    }
    Some(execute(&mut cmd, args, false))
}

pub fn run_js(args: &Args) -> Option<i32> {
    let mut cmd = Command::new(args.prog("node"));
    cmd.arg(&args.file_names[0]).args(&args.exec_args);
    if args.watch {
        return Some(run_watch(args, cmd, None, &[]));
    }
    Some(execute(&mut cmd, args, false))
}

pub fn run_ml(args: &Args) -> Option<i32> {
    let prog = args.prog("ocamlopt");
    let exts: &[&str] = if !args.keep_all_out {
        if prog == "ocamlopt" {
            &["cmi", "cmx", "o"]
        } else if prog == "ocamlc" {
            &["cmo", "cmi"]
        } else {
            &[]
        }
    } else {
        &[]
    };
    let mut cmd = new_compile_cmd(&prog, args);
    let status = execute(&mut cmd, args, true);
    if exts.len() != 0 {
        delete_files(exts, args);
    }
    if status == 0 {
        None
    } else {
        Some(status)
    }
}

pub fn run_pl(args: &Args) -> Option<i32> {
    let mut cmd = Command::new(args.prog("perl"));
    cmd.arg(&args.file_names[0]).args(&args.exec_args);
    if args.watch {
        return Some(run_watch(args, cmd, None, &[]));
    }
    Some(execute(&mut cmd, args, false))
}

pub fn run_py(args: &Args) -> Option<i32> {
    let mut cmd = Command::new(args.prog("python3"));
    cmd.arg(&args.file_names[0]).args(&args.exec_args);
    if args.watch {
        return Some(run_watch(args, cmd, None, &[]));
    }
    Some(execute(&mut cmd, args, false))
}

pub fn run_r(args: &Args) -> Option<i32> {
    let mut cmd = Command::new(args.prog("Rscript"));
    cmd.arg(&args.file_names[0]).args(&args.exec_args);
    if args.watch {
        return Some(run_watch(args, cmd, None, &[]));
    }
    Some(execute(&mut cmd, args, false))
}

pub fn run_rs(args: &Args) -> Option<i32> {
    if args.file_names.len() == 0 {
        let mut cmd = Command::new(args.prog("cargo"));
        // TODO: Properly time the execution of the program.
        if args.compile_only {
            cmd.arg("build").args(&args.comp_args);
        } else {
            cmd.arg("run").args(&args.comp_args);
            if args.exec_args.len() != 0 {
                cmd.arg("--").args(&args.exec_args);
            }
        }
        if args.watch {
            return Some(run_watch(args, cmd, None, &[]));
        }
        return Some(execute(&mut cmd.args(&args.comp_args), args, !args.no_out));
    }
    let mut cmd = new_compile_cmd(args.prog("rustc"), args);
    if args.default_flags {
        cmd.arg("--edition=2021");
    }
    if args.watch {
        return Some(run_watch(args, new_exec_cmd(args), Some(cmd), &[]));
    }
    let status = execute(&mut cmd, args, true);
    if status == 0 {
        None
    } else {
        Some(status)
    }
}

pub fn run_swift(args: &Args) -> Option<i32> {
    // TODO: Properly format swift args
    if args.no_out {
        let mut cmd = Command::new(args.prog("swift"));
        cmd.arg(&args.file_names[0]).args(&args.exec_args);
        if args.watch {
            return Some(run_watch(args, cmd, None, &[]));
        }
        return Some(execute(&mut cmd, args, false));
    }
    let mut cmd = new_compile_cmd(&args.prog("swiftc"), args);
    if args.watch {
        return Some(run_watch(args, new_exec_cmd(args), Some(cmd), &[]));
    }
    let status = execute(&mut cmd, args, true);
    if status == 0 {
        None
    } else {
        Some(status)
    }
}

pub fn run_executable(args: &Args) -> i32 {
    execute(&mut new_exec_cmd(args), args, false)
}

fn new_exec_cmd(args: &Args) -> Command {
    match args.file_type.unwrap() {
        FileType::JAV => {
            let mut cmd = Command::new("java");
            cmd.arg(&args.file_names[0].with_extension(""))
                .args(&args.exec_args);
            return cmd;
        }
        _ => (),
    }
    let name = if let Some(name) = &args.output_name {
        if format!("{}", name.parent().unwrap().display()).len() != 0 {
            name.clone()
        } else {
            std::path::PathBuf::from(".").join(name)
        }
    } else {
        die!(2, "output name not set")
    };
    let mut cmd = Command::new(name);
    cmd.args(&args.exec_args);
    cmd
}

fn execute(cmd: &mut Command, args: &Args, compiling: bool) -> i32 {
    let start = if !args.no_time {
        print_start(cmd, compiling);
        // Start timing
        Some(Instant::now())
    } else {
        None
    };
    let status = cmd
        .status()
        .unwrap_or_else(|err| die!(3, "error encountered: {}", err));
    let code = status.code().unwrap_or_else(|| {
        // TODO: Possibly use different default than -1
        if cfg!(unix) {
            use std::os::unix::process::ExitStatusExt;
            status.signal().unwrap_or(-1)
        } else {
            -1
        }
    });
    // If the command was timed
    if let Some(start) = start {
        print_end_time(start, compiling);
    }
    code
}

fn cmd_string(cmd: &Command) -> String {
    cmd.get_args()
        .fold(cmd.get_program().to_string_lossy().to_string(), |s, arg| {
            s + " " + &arg.to_string_lossy()
        })
}

fn print_start(cmd: &Command, compiling: bool) {
    println!(
        "{}...\t{}\n",
        if compiling { "Compiling" } else { "Executing" },
        cmd_string(cmd)
    );
}

fn print_end_time(start: Instant, compiling: bool) {
    println!(
        "\n{} time: {} seconds",
        if compiling {
            "Compilation"
        } else {
            "Execution"
        },
        start.elapsed().as_secs_f64(),
    );
}

fn new_compile_cmd<S: ToString>(compiler: S, args: &Args) -> Command {
    // TODO: Handle Better
    let mut cmd = Command::new(compiler.to_string());
    cmd.arg("-o")
        .arg(
            args.output_name
                .as_ref()
                .unwrap_or_else(|| die!(2, "output name not set")),
        )
        .args(&args.file_names)
        .args(&args.comp_args);
    cmd
}

fn delete_files(exts: &[&str], args: &Args) {
    let base_name = &args.file_names[0];
    for &ext in exts.iter() {
        remove_file(&base_name.with_extension(ext)).unwrap_or_else(|err| {
            eprintln!("error deleting intermediate output file: {}", err);
        });
    }
}

pub fn run_watch(
    args: &Args,
    mut exec_cmd: Command,
    mut comp_cmd: Option<Command>,
    exts_to_delete: &[&str],
) -> i32 {
    // Used a a noop for the delete f unc
    fn no_delete(_: &[&str], _: &Args) {}
    use crate::watcher::*;
    let rx = match watch(&args.file_names[0]) {
        Ok(rx) => rx,
        Err(e) => die!(1, "error starting watcher: {}", e),
    };
    let delete_func = if exts_to_delete.len() != 0 {
        delete_files
    } else {
        no_delete
    };
    // TODO: Time
    let mut child = None::<Child>;
    if let Some(cmd) = comp_cmd.as_ref() {
        println!("Compilation command: {}", cmd_string(cmd));
    }
    println!("Execution command: {}", cmd_string(&exec_cmd));
    let padding = "=".repeat(20);
    for event in rx {
        if let Err(e) = child.take().map(|mut c| c.kill()).transpose() {
            println!("\n{0}Execution complete{0}\n", padding);
            // Invalid input is returned if the child has already exited
            if e.kind() != std::io::ErrorKind::InvalidInput {
                die!(3, "error encountered: {}", e);
            }
        }
        match event {
            FileEvent::Write => (),
            FileEvent::Remove => die!(0, "file removed, stopping..."),
            FileEvent::Error(e) => die!(1, "error watching: {}", e),
        }
        if let Some(cmd) = comp_cmd.as_mut() {
            let start = (!args.no_time).then(|| Instant::now());
            println!("\n{0}Compiling...{0}\n", padding);
            if !cmd
                .status()
                .unwrap_or_else(|e| die!(3, "error encountered: {}", e))
                .success()
            {
                continue;
            }
            if let Some(elapsed) = start.map(|t| t.elapsed().as_secs_f64()) {
                println!("\n{1}Compilation time: {0:.3}{1}\n", elapsed, padding);
            }
            delete_func(exts_to_delete, args);
        }
        println!("\n{0}Executing...{0}\n", padding);
        child = Some(
            exec_cmd
                .spawn()
                .unwrap_or_else(|e| die!(3, "error encountered: {}", e)),
        );
    }
    0
}

pub fn run_multiple(args: &Args) -> i32 {
    crate::app::run_app(args)
}
