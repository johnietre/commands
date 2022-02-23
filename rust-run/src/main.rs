use clap::Parser;
use std::process::{exit, Command};
use std::time::Instant;

#[derive(Parser, Debug)]
struct Args {
    file_name: String,

    #[clap(short, long)]
    output_name: Option<String>,

    #[clap(short, long)]
    compile_only: bool,

    #[clap(short, long)]
    delete: bool,

    #[clap(long)]
    comp_arg: Vec<String>,

    #[clap(long)]
    exec_arg: Vec<String>,

    #[clap(long, parse(try_from_str))]
    file_type: Option<FileType>,

    #[clap(long)]
    no_time: bool,

    #[clap(long)]
    keep_hs: bool,

    #[clap(short, long)]
    wasm: bool,

    #[clap(long)]
    parse_includes: bool,
}

#[derive(Debug)]
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
    PL, // TODO: Specify between perl and prolog
    PY,
    R,
    RS,
    SWIFT,
}

impl std::str::FromStr for FileType {
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
    let args = Args::parse();
    execute(Command::new("run").arg("main.py"), &args, false);
}

fn execute(cmd: &mut Command, args: &Args, compiling: bool) {
    if !args.no_time {
        println!(
            "{}...\t{}\n",
            if compiling { "Compiling" } else { "Executing" },
            cmd.get_program().to_string_lossy().to_string() + &cmd.get_args()
                .map(|s| s.to_string_lossy().to_string())
                .collect::<Vec<String>>()
                .join(" ")
                //.fold(cmd.get_program().to_string_lossy().to_string(), |s, arg| {
                    //s + &arg.to_string_lossy()
                //})
        );
    }
    // Start timing and run the command
    let start = Instant::now();
    let status = cmd.status().unwrap_or_else(|err| {
        eprintln!("error encountered: {}", err);
        exit(1)
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
}
