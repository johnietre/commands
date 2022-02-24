// TODO: Clean and refactor
// Add help
use clap::Parser;
use std::collections::hash_map::DefaultHasher;
use std::env::temp_dir;
use std::fs::{canonicalize, remove_file};
use std::hash::{Hash, Hasher};
use std::path::PathBuf;
use std::process::{exit, Command};
use std::str::FromStr;
use std::time::Instant;

#[derive(Parser, Debug)]
struct Args {
    file_names: Vec<PathBuf>,

    #[clap(short, long = "output")]
    output_name: Option<PathBuf>,

    #[clap(short, long)]
    compile_only: bool,

    #[clap(long = "comp_arg")]
    comp_args: Vec<String>,

    #[clap(long = "exec_arg")]
    exec_args: Vec<String>,

    #[clap(short = 't', long = "type", parse(try_from_str))]
    file_type: Option<FileType>,

    #[clap(short, long)]
    no_out: bool,

    #[clap(long)]
    temp: bool,

    #[clap(short, long)]
    default_flags: bool,

    #[clap(long)]
    no_time: bool,

    #[clap(long)]
    keep_all_out: bool,

    #[clap(short, long)]
    wasm: bool,

    #[clap(long)]
    program: Option<String>,

    #[clap(long)]
    parse_includes: bool,
}

impl Args {
    fn prog<S: ToString>(&self, alt: S) -> String {
        self.program.as_ref().cloned().unwrap_or(alt.to_string())
    }
}

#[derive(Debug, Clone, Copy, PartialEq)]
enum FileType {
    C,
    CPP,
    F90,
    GO,
    HS,
    JAV,
    JL,
    JS,
    ML,
    PL,
    PY,
    R,
    RS,
    SWIFT,
}

impl FileType {
    fn can_have_multiple(self) -> bool {
        use FileType::*;
        match self {
            C | CPP | GO | JAV => true,
            _ => false,
        }
    }

    fn can_have_none(self) -> bool {
        use FileType::*;
        match self {
            GO | RS => true,
            _ => false,
        }
    }
}

impl FromStr for FileType {
    type Err = &'static str;
    fn from_str(s: &str) -> Result<Self, Self::Err> {
        use FileType::*;
        match s {
            "c" => Ok(C),
            "cpp" | "cc" | "cxx" => Ok(CPP),
            "f77" | "f90" | "f95" => Ok(F90),
            "go" => Ok(GO),
            "hs" => Ok(HS),
            "jav" | "java" => Ok(JAV),
            "jl" => Ok(JL),
            "js" => Ok(JS),
            "ml" => Ok(ML),
            "pl" | "perl" => Ok(PL),
            "py" => Ok(PY),
            "r" | "R" | "Rscript" => Ok(R),
            "rs" => Ok(RS),
            "swift" => Ok(SWIFT),
            _ => Err("invalid file type"),
        }
    }
}

fn main() {
    use FileType::*;
    let mut args = Args::parse();
    // Get the file type
    let file_type = if let Some(t) = args.file_type {
        if args.file_names.len() == 0 && !t.can_have_none() {
            eprintln!("must include file name");
            exit(1);
        }
        t
    } else {
        let file_name = args.file_names.get(0).map_or_else(
            || {
                eprintln!("must include file name");
                exit(1)
            },
            |name| name,
        );
        FileType::from_str(&file_name.extension().unwrap_or_default().to_string_lossy())
            .unwrap_or_else(|err| {
                eprintln!("{}", err);
                exit(1)
            })
    };
    // Check the file name(s)
    if args.file_names.len() > 1 && !file_type.can_have_multiple() {
        eprintln!("only one file name allowed");
        exit(1);
    }
    args.file_type.replace(file_type);
    if args.no_out {
        args.temp = true;
    }
    // Get the output name
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
                name.set_extension(".exe");
            }
            args.output_name.replace(name);
        }
    }
    if args.wasm {
        args.compile_only = true;
    }
    // Get the func to run the file
    let run_func = match file_type {
        C => run_c,
        CPP => run_cpp,
        F90 => run_f90,
        GO => run_go,
        HS => run_hs,
        JAV => run_jav,
        JL => run_jl,
        JS => run_js,
        ML => run_ml,
        PL => run_pl,
        PY => run_py,
        R => run_r,
        RS => run_rs,
        SWIFT => run_swift,
    };
    // Run the compilation or execution
    let status = run_func(&args);
    // If the file was compiled and needs to be run, run the executable,
    // otherwise, exit
    if args.compile_only || status != 0 {
        exit(status)
    } else {
        exit(run_executable(&args))
    }
}

fn run_c(args: &Args) -> i32 {
    let defaults = if args.default_flags {
        vec!["-std=c18"]
    } else {
        vec![]
    };
    execute(
        new_comp_cmd(args.prog("gcc"), args).args(&defaults),
        args,
        true,
    )
}

fn run_cpp(args: &Args) -> i32 {
    let defaults = if args.default_flags {
        vec!["-std=gnu++17"]
    } else {
        vec![]
    };
    execute(
        new_comp_cmd(args.prog("g++"), args).args(&defaults),
        args,
        true,
    )
}

fn run_f90(args: &Args) -> i32 {
    execute(&mut new_comp_cmd(args.prog("gfortran"), args), args, true)
}

fn run_go(args: &Args) -> i32 {
    if args.no_out && !args.temp {
        exit(execute(
            Command::new("go")
                .arg("run")
                .args(&args.comp_args)
                .args(&args.file_names)
                .args(&args.exec_args),
            args,
            false,
        ));
    }
    let envs = if args.wasm {
        vec![("GOOS", "js"), ("GOARCH", "wasm")]
    } else {
        vec![]
    };
    execute(
        Command::new("go")
            .envs(envs)
            .arg("build")
            .arg("-o")
            .arg(args.output_name.as_ref().unwrap_or_else(|| {
                eprintln!("output name not set");
                exit(1);
            }))
            .args(&args.comp_args)
            .args(&args.file_names),
        args,
        true,
    )
}

fn run_hs(args: &Args) -> i32 {
    if args.no_out && !args.temp {
        exit(execute(
            Command::new(args.prog("runghc"))
                .arg(&args.file_names[0])
                .args(&args.exec_args),
            args,
            false,
        ));
    }
    let prog = args.prog("ghc");
    let status = execute(&mut new_comp_cmd(&prog, args), args, true);
    if !args.keep_all_out && prog == "ghc" {
        delete_files(&["o", "hi"], &args.file_names[0]);
    }
    status
}

fn run_jav(args: &Args) -> i32 {
    // TODO: Properly format javac arguments
    execute(
        Command::new("javac")
            .args(&args.file_names)
            .args(&args.comp_args),
        args,
        true,
    )
}

fn run_jl(args: &Args) -> i32 {
    exit(execute(
        Command::new(args.prog("julia"))
            .arg(&args.file_names[0])
            .args(&args.exec_args),
        args,
        false,
    ))
}

fn run_js(args: &Args) -> i32 {
    exit(execute(
        Command::new(args.prog("node"))
            .arg(&args.file_names[0])
            .args(&args.exec_args),
        args,
        false,
    ))
}

fn run_ml(args: &Args) -> i32 {
    let prog = args.prog("ocamlopt");
    let status = execute(&mut new_comp_cmd(&prog, args), args, true);
    if !args.keep_all_out {
        let exts = if prog == "ocamlopt" {
            vec!["cmi", "cmx", "o"]
        } else if prog == "ocamlc" {
            vec!["cmo", "cmi"]
        } else {
            vec![]
        };
        delete_files(&exts, &args.file_names[0]);
    }
    status
}

fn run_pl(args: &Args) -> i32 {
    exit(execute(
        Command::new(args.prog("perl"))
            .arg(&args.file_names[0])
            .args(&args.exec_args),
        args,
        false,
    ))
}

fn run_py(args: &Args) -> i32 {
    exit(execute(
        Command::new(args.prog("python3"))
            .arg(&args.file_names[0])
            .args(&args.exec_args),
        args,
        false,
    ))
}

fn run_r(args: &Args) -> i32 {
    exit(execute(
        Command::new(args.prog("Rscript"))
            .arg(&args.file_names[0])
            .args(&args.exec_args),
        args,
        false,
    ))
}

fn run_rs(args: &Args) -> i32 {
    if args.file_names.len() == 0 {
        let mut cmd = Command::new(args.prog("cargo"));
        if args.no_out {
            cmd.arg("run");
        } else {
            cmd.arg("build");
        }
        exit(execute(cmd.args(&args.comp_args), args, !args.no_out));
    }
    execute(&mut new_comp_cmd(args.prog("rustc"), args), args, true)
}

fn run_swift(args: &Args) -> i32 {
    // TODO: Properly format swift args
    if args.no_out {
        exit(execute(
            Command::new(args.prog("swift"))
                .arg(&args.file_names[0])
                .args(&args.exec_args),
            args,
            false,
        ));
    }
    execute(&mut new_comp_cmd(&args.prog("swiftc"), args), args, true)
}

fn run_executable(args: &Args) -> i32 {
    if args.file_type.unwrap() == FileType::JAV {
        return execute(
            Command::new("java")
                .arg(&args.file_names[0].with_extension(""))
                .args(&args.exec_args),
            args,
            false,
        );
    }
    let name = if let Some(name) = &args.output_name {
        if format!("{}", name.parent().unwrap().display()).len() != 0 {
            name.clone()
        } else {
            PathBuf::from(".").join(name)
        }
    } else {
        eprintln!("output name not set");
        exit(1)
    };
    execute(Command::new(name).args(&args.exec_args), args, false)
}

fn execute(cmd: &mut Command, args: &Args, compiling: bool) -> i32 {
    if !args.no_time {
        println!(
            "{}...\t{}\n",
            if compiling { "Compiling" } else { "Executing" },
            cmd.get_args()
                .fold(cmd.get_program().to_string_lossy().to_string(), |s, arg| {
                    s + " " + &arg.to_string_lossy()
                })
        );
    }
    // Start timing and run the command
    let start = Instant::now();
    let status = cmd.status().unwrap_or_else(|err| {
        eprintln!("error encountered: {}", err);
        exit(1)
    });
    let code = status.code().unwrap_or_else(|| {
        // TODO: Possibly use different default than -1
        if cfg!(unix) {
            use std::os::unix::process::ExitStatusExt;
            status.signal().unwrap_or(-1)
        } else {
            -1
        }
    });
    // Get the time elapsed
    let elapsed = start.elapsed();
    if !args.no_time {
        println!(
            "\n{} time: {} seconds",
            if compiling {
                "Compilation"
            } else {
                "Execution"
            },
            elapsed.as_secs_f64(),
        );
    }
    code
}

fn new_comp_cmd<S: ToString>(compiler: S, args: &Args) -> Command {
    // TODO: Handle Better
    let mut cmd = Command::new(compiler.to_string());
    cmd.arg("-o")
        .arg(args.output_name.as_ref().unwrap_or_else(|| {
            eprintln!("output name not set");
            exit(1);
        }))
        .args(&args.file_names)
        .args(&args.comp_args);
    cmd
}

fn delete_files(exts: &[&str], base_name: &PathBuf) {
    for &ext in exts.iter() {
        remove_file(&base_name.with_extension(ext)).unwrap_or_else(|err| {
            eprintln!("error deleting intermediate output file: {}", err);
        });
    }
}
