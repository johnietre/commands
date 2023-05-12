// TODO: Clean and refactor
// TODO: Add "--bash" and "-b" flags to be passed to run a program as bash
// TODO: Allow flags to be passed to programs
// TODO: Delete output after --no-out flag
// TODO: Stuff like segmentation fault message not being output
// Building go directory not working
// In order to pass flags through comp or exec flag, use: --[comp/exec]_arg=--flag
// To pass multiple arguments in one call, use: --[comp/exec]_arg={--flag1,arg1,-opt}
// Add help
use clap::Parser;
use libc::{read, tcsetattr, termios, TCSANOW};
use std::collections::{hash_map::DefaultHasher, LinkedList};
use std::env::temp_dir;
use std::fs::{canonicalize, remove_file};
use std::hash::{Hash, Hasher};
use std::io::{self, Error, ErrorKind, Write};
use std::os::unix::io::AsRawFd;
use std::path::PathBuf;
use std::process::{exit, Command, ExitStatus};
use std::str::FromStr;
use std::sync::{
    atomic::{AtomicBool, Ordering},
    Arc, Mutex,
};
use std::time::Instant;

mod ansi;
use ansi::*;

macro_rules! die {
    ($code:expr, $($args:tt)*) => ({
        eprintln!($($args)*);
        ::std::process::exit($code)
    });
}

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

    #[clap(short, long)]
    bash: bool,
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
        match s.to_lowercase().as_str() {
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
            unsafe { exit(run_multiple(&args)) }
            /*
            eprintln!("must include file name");
            exit(1);
            */
        }
        t
    } else {
        let file_name = args
            .file_names
            .get(0)
            .map_or_else(|| unsafe { exit(run_multiple(&args)) }, |name| name);
        FileType::from_str(&file_name.extension().unwrap_or_default().to_string_lossy())
            .unwrap_or_else(|err| die!(1, "error parsing file type: {}", err))
    };
    // Check the file name(s)
    if args.file_names.len() > 1 && !file_type.can_have_multiple() {
        die!(1, "only one file name allowed")
    }
    args.file_type.replace(file_type);
    if args.no_out {
        //args.temp = true;
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
    //if args.no_out { TODO: Make this line active and one below not
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
            .arg(
                args.output_name
                    .as_ref()
                    .unwrap_or_else(|| die!(2, "output name not set")),
            )
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
        delete_files(&["o", "hi"], args);
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
        delete_files(&exts, args);
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
        die!(2, "output name not set")
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
        .arg(
            args.output_name
                .as_ref()
                .unwrap_or_else(|| die!(2, "output name not set")),
        )
        .args(&args.file_names)
        .args(&args.comp_args);
    cmd
}

fn delete_files(exts: &[&str], base_name: &Args) {
    for &ext in exts.iter() {
        remove_file(&base_name.with_extension(ext)).unwrap_or_else(|err| {
            eprintln!("error deleting intermediate output file: {}", err);
        });
    }
}

unsafe fn run_multiple(_args: &Args) -> i32 {
    //let mut deferrer = Rc::new(RefCell::new(Deferrer::new()));
    let deferrer = Arc::new(Mutex::new(Deferrer::new()));
    let mut stdout = io::stdout().lock();
    let stdin_fd = io::stdin().as_raw_fd();
    // Get the old terminal and create the new
    let mut old_term = termios {
        c_iflag: 0,
        c_oflag: 0,
        c_cflag: 0,
        c_lflag: 0,
        c_line: 0,
        c_cc: [0; 32],
        c_ispeed: 0,
        c_ospeed: 0,
    };
    if libc::tcgetattr(stdin_fd, (&mut old_term) as *mut _) != 0 {
        die!(3, "error getting terminal info");
    }
    // Set panic handler
    let df = Arc::downgrade(&deferrer);
    std::panic::set_hook(Box::new(move |info| {
        eprintln!(
            "panic occurred: {}",
            info.payload().downcast_ref::<&str>().unwrap_or(&"")
        );
        if let Some(df) = df.upgrade() {
            df.lock().unwrap().force_run();
        }
    }));
    // Set ctrlc handler
    let running = Arc::new(AtomicBool::new(true));
    let (r, df) = (Arc::clone(&running), Arc::downgrade(&deferrer));
    ctrlc::set_handler(move || {
        // Second ctrl-c force quits
        if !r.swap(false, Ordering::SeqCst) {
            if let Some(df) = df.upgrade() {
                df.lock().unwrap().force_run();
            }
            exit(0);
        }
    })
    .unwrap_or_else(|_| die!(3, "error setting Ctrl-C handler"));
    // Set the terminal
    let new_term = termios {
        c_lflag: old_term.c_lflag & (!libc::ICANON & !libc::ECHO),
        c_cc: {
            let mut vals = old_term.c_cc;
            vals[libc::VMIN] = 0; // Minimum number of btyes required to be read
            vals[libc::VTIME] = 1; // Amount of time to wait before read returns
            vals
        },
        ..old_term
    };
    if tcsetattr(stdin_fd, TCSANOW, (&new_term) as *const _) != 0 {
        die!(3, "error setting terminal info");
    }
    deferrer.lock().unwrap().push_back(Box::new(move || {
        if tcsetattr(stdin_fd, TCSANOW, (&old_term) as *const _) != 0 {
            die!(3, "error resetting terminal info");
        }
    }));
    // TODO: Fix Deferrer; not going to run due

    print!("{}{}{}", SCREEN_SAVE, CLEAR_SCREEN, CUR_TO_HOME);
    stdout.flush().unwrap();
    let restore_func = deferrer.lock().unwrap().push_front(Box::new(move || {
        // TODO: Figure out better way to handle
        print!("{}{}{}", CLEAR_SCREEN, CUR_TO_HOME, SCREEN_RESTORE);
    }));

    let mut counter = 0;
    let mut children = LinkedList::new();
    let mut changed = Change::NA;
    let mut buf = [0 as u8; 3];
    // 1-based position
    let mut pos = Pos { x: 1, y: 1 };
    while running.load(Ordering::SeqCst) {
        if changed != Change::NA {
            let (skip, y, end) = match changed {
                Change::Add => (
                    children.len(),
                    children.len(),
                    format!("\r{}{}", children.back().unwrap(), CUR_TO_POS(1, pos.y)),
                ),
                Change::Del(y) => (y - 1, y, format!("{}{}", CLEAR_LINE, CUR_TO_POS(1, pos.y))),
                Change::NA => unreachable!(),
            };
            let changes = children
                .iter()
                .skip(skip)
                .fold(CUR_TO_POS(1, y), |acc, &v| {
                    acc + &format!("\r{}{}{}", CLEAR_LINE, v, CUR_DOWN)
                })
                + &end;
            print!("{}", changes);
            stdout.flush().unwrap();
            changed = Change::NA;
        }
        let n = read(stdin_fd, buf.as_mut_ptr().cast(), 3);
        if n == -1 {
            eprintln!("error reading from stdin, quitting");
            running.store(false, Ordering::SeqCst);
            break;
        } else if n == 0 {
            continue;
        }
        if n == 1 {
            match buf[0] {
                b'q' => {
                    running.store(false, Ordering::SeqCst);
                    break;
                }
                b'a' => {
                    counter += 1;
                    children.push_back(counter);
                    changed = Change::Add;
                }
                b'd' => {
                    if children.len() != 0 {
                        children = children
                            .into_iter()
                            .enumerate()
                            .filter(|(i, _)| *i != pos.y - 1)
                            .map(|(_, v)| v)
                            .collect();
                        changed = Change::Del(pos.y);
                        if pos.y != 1 {
                            pos.y -= 1;
                            print!("{}", CUR_TO_POS(1, pos.y));
                            stdout.flush().unwrap();
                        }
                    }
                }
                _ => panic!("{}", buf[0] as char),
            }
        } else if n == 2 {
            panic!("2");
        } else {
            // Arrow key
            if buf[0] == 27 && buf[1] == 91 {
                match buf[2] {
                    65 => {
                        if pos.y != 1 {
                            pos.y -= 1;
                            print!("{}", CUR_UP);
                            stdout.flush().unwrap();
                        }
                    }
                    66 => {
                        // Down
                        if children.len() != 0 && pos.y != children.len() {
                            pos.y += 1;
                            print!("{}", CUR_DOWN);
                            stdout.flush().unwrap();
                        }
                    }
                    67 => (), // Right
                    68 => (), // Right
                    _ => (),  // Unknown/unhandled sequence
                }
            } else {
                panic!("3");
            }
        }
    }
    restore_func.should_run(false);
    if children.len() == 0 {
        print!("{}{}", CUR_TO_HOME, SCREEN_RESTORE);
    } else {
        print!("{}\n{}", CUR_TO_POS(1, children.len()), SCREEN_RESTORE);
    };
    for _child in children.iter() {
        //libc::kill(child.id() as _, libc::SIGINT);
    }
    0
}

#[derive(Clone, Copy, PartialEq, Eq, Default)]
struct Pos {
    x: usize,
    y: usize,
}

#[derive(Clone, Copy, PartialEq, Eq)]
enum Change {
    NA,
    Add,
    Del(usize),
}

#[derive(Clone, Copy, PartialEq, Eq)]
enum ProcessStatus {
    NotStarted,
    Running,
    Finished(ExitStatus),
}

struct Process {
    name: String,
    cmd: Command,
    child: Option<std::process::Child>,
    status: ProcessStatus,
}

#[allow(dead_code)]
impl Process {
    fn new(name: String, cmd: Command) -> Self {
        Self {
            name,
            cmd,
            child: None,
            status: ProcessStatus::NotStarted,
        }
    }

    fn start(&mut self) -> io::Result<()> {
        if self.is_alive() {
            return Err(Error::new(
                ErrorKind::InvalidInput,
                "process is already alive",
            ));
        }
        let child = self.cmd.spawn()?;
        self.child.replace(child);
        self.status = ProcessStatus::Running;
        Ok(())
    }

    fn kill(&mut self) -> io::Result<()> {
        if self.is_alive() {
            self.child.as_mut().expect("child is none").kill()
        } else {
            Err(Error::new(ErrorKind::InvalidInput, "process isn't alive"))
        }
    }

    fn interrupt(&mut self) -> io::Result<()> {
        if self.is_alive() {
            unsafe {
                libc::kill(
                    self.child.as_ref().expect("child is none").id() as _,
                    libc::SIGINT,
                )
            };
            Ok(())
        } else {
            Err(Error::new(ErrorKind::InvalidInput, "process isn't alive"))
        }
    }

    fn signal(&mut self, sig: libc::c_int) {
        // TODO: Return some kind of error?
        unsafe { libc::kill(self.child.as_ref().expect("child is none").id() as _, sig) };
    }

    fn is_alive(&self) -> bool {
        use ProcessStatus::*;
        match self.status {
            Running => true,
            NotStarted | Finished(_) => false,
        }
    }

    fn name(&self) -> &str {
        &self.name
    }

    fn id(&self) -> Option<u32> {
        Some(self.child.as_ref()?.id())
    }

    fn try_wait(&mut self) -> io::Result<Option<ExitStatus>> {
        if let ProcessStatus::Finished(status) = self.status {
            return Ok(Some(status));
        }
        Ok(
            match self.child.as_mut().expect("child is none").try_wait()? {
                Some(status) => {
                    self.child.take();
                    self.status = ProcessStatus::Finished(status);
                    Some(status)
                }
                None => None,
            },
        )
    }

    fn wait(&mut self) -> io::Result<ExitStatus> {
        if let ProcessStatus::Finished(status) = self.status {
            return Ok(status);
        }
        let status = self.child.take().expect("child is none").wait()?;
        self.status = ProcessStatus::Finished(status);
        Ok(status)
    }
}

type DeferFunc = Box<dyn Fn() + Send + Sync + 'static>;

struct Defer {
    f: DeferFunc,
    should_run: Arc<AtomicBool>,
}

impl Defer {
    fn new(f: DeferFunc) -> Self {
        Self {
            should_run: Arc::new(AtomicBool::new(true)),
            f,
        }
    }

    fn run(&self) {
        if self.should_run.load(Ordering::SeqCst) {
            (self.f)();
        }
    }

    fn should_run(&self, b: bool) {
        self.should_run.store(b, Ordering::SeqCst);
    }
}

// The bool tells whether it has executed or not
// In order to still run when original function scope ends but it's in multiple threads,
// wrap in an Arc and use Weak for the threads
#[derive(Default)]
struct Deferrer(LinkedList<Arc<Defer>>, bool);

impl Deferrer {
    const fn new() -> Self {
        Self(LinkedList::new(), false)
    }

    fn push_back(&mut self, f: DeferFunc) -> Arc<Defer> {
        let f = Arc::new(Defer::new(f));
        self.0.push_back(Arc::clone(&f));
        f
    }

    fn push_front(&mut self, f: DeferFunc) -> Arc<Defer> {
        let f = Arc::new(Defer::new(f));
        self.0.push_front(Arc::clone(&f));
        f
    }

    fn force_run(&mut self) {
        if !self.1 {
            self.0.iter().for_each(|d| d.run());
            self.1 = true;
        }
    }
}

impl Drop for Deferrer {
    fn drop(&mut self) {
        self.force_run();
    }
}
