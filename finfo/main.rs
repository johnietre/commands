use std::path::{Path, PathBuf};
use std::sync::{mpsc::{self, Sender, channel}, Arc, Condvar, Mutex};
use std::thread;

macro_rules! die {
    ($code:expr, $($args:tt)*) => {{
        ::std::eprintln!($($args)*);
        ::std::process::exit($code)
    }}
}

static POOL: ThreadPool = ThreadPool::new_empty();

fn main() {
    let args = parse_args();
    run(args);
}

struct Args {
    sort: bool,
    threads: usize,
    paths: Vec<PathBuf>,
}

impl Default for Args {
    fn default() -> Self {
        Self {
            sort: false,
            threads: thread::available_parallelism().map_or(1, |nz| nz.into()),
            paths: Vec::new(),
        }
    }
}

struct Finfo {
    path: PathBuf,
    size: u64,
}

impl Finfo {
    fn new(path: PathBuf, size: u64) -> Self {
        Self { path, size }
    }
}

fn run(args: Args) {
    let args = Arc::new(args);
    POOL.start_workers(args.threads);

    let (tx, rx) = channel();
    for path in &args.paths {
        let path = path.clone();
        let tx = tx.clone();
        POOL.submit_job(Box::new(move || walk(path, tx)));
    }
    drop(tx);

    let mut finfos = Vec::new();
    for (i, finfo) in rx.into_iter().enumerate() {
        if args.sort {
            finfos.push(finfo);
        } else {
            if i != 0 {
                println!("{}", "-".repeat(40));
            }
            println!("{}: {}", finfo.path.display(), make_size_string(finfo.size));
        }
    }
    POOL.wait();
    finfos.sort_by_cached_key(|fi| (fi.size, fi.path.clone()));
    for (i, finfo) in finfos.into_iter().enumerate() {
        if i != 0 {
            println!("{}", "-".repeat(40));
        }
        println!("{}: {}", finfo.path.display(), make_size_string(finfo.size));
    }
}

fn walk<P: AsRef<Path>>(path: P, tx: Sender<Finfo>) {
    let path = path.as_ref();
    let ents = match path.read_dir() {
        Ok(ents) => ents,
        Err(e) => {
            if e.to_string().contains("Too many open files") {
                let path = path.to_path_buf();
                POOL.submit_job(Box::new(move || walk(path, tx)));
            } else {
                eprintln!("Error reading {}: {e}", path.display());
            }
            return;
        }
    };
    for ent in ents {
        let ent = match ent {
            Ok(ent) => ent,
            Err(e) => {
                eprintln!("Error reading entry in {}: {e}", path.display());
                continue;
            }
        };
        let md = match ent.metadata() {
            Ok(md) => md,
            Err(e) => {
                eprintln!("Error reading metadata for {}: {e}", ent.path().display());
                continue;
            }
        };
        if md.is_symlink() {
        } else if md.is_dir() {
            let tx = tx.clone();
            POOL.submit_job(Box::new(move || walk(ent.path().clone(), tx)));
        } else {
            tx.send(Finfo::new(ent.path().clone(), md.len())).unwrap();
        }
    }
}

type TPFn = Box<dyn FnOnce() + 'static + Send + Sync>;

struct ThreadPool {
    workers: Mutex<Vec<thread::JoinHandle<()>>>,
    tx: Mutex<Option<Sender<TPFn>>>,
    done: Mutex<bool>,
    done_cv: Condvar,
}

impl ThreadPool {
    const fn new_empty() -> Self {
        Self {
            workers: Mutex::new(Vec::new()),
            tx: Mutex::new(None),
            done: Mutex::new(true),
            done_cv: Condvar::new(),
        }
    }

    fn new(n_workers: usize) -> Self {
        let pool = Self::new_empty();
        pool.start_workers(n_workers);
        pool
    }
    
    fn start_workers(&self, n_workers: usize) -> bool {
        let txo = self.tx.lock().unwrap();
        if txo.is_none() {
            drop(txo);
            self.wait();
        } else {
            return false;
        }
        let (tx, rx) = mpsc::channel::<TPFn>();
        let amrx = Arc::new(Mutex::new(rx));
        let mut workers = Vec::with_capacity(n_workers);
        for _ in 0..n_workers {
            let amrx = Arc::clone(&amrx);
            let handle = thread::spawn(move || {
                loop {
                    let f = {
                        let rx = amrx.lock().unwrap();
                        match rx.recv() {
                            Ok(f) => f,
                            Err(_) => return,
                        }
                    };
                    f();
                }
            });
            workers.push(handle);
        }
        *self.workers.lock().unwrap() = workers;
        *self.tx.lock().unwrap() = Some(tx);
        *self.done.lock().unwrap() = false;
        true
    }

    fn submit_job(&self, f: TPFn) {
        if let Some(tx) = self.tx.lock().unwrap().as_mut() {
            let _ = tx.send(f);
        }
    }

    #[inline(always)]
    fn n_workers(&self) -> usize {
        self.workers.lock().unwrap().len()
    }

    fn wait(&self) {
        let mut txo = self.tx.lock().unwrap();
        if txo.take().is_none() {
            drop(txo);
            let mut done = self.done.lock().unwrap();
            while !*done {
                done = self.done_cv.wait(done).unwrap();
            }
            return;
        }
        let mut workers = self.workers.lock().unwrap();
        for worker in workers.drain(..) {
            let _ = worker.join();
        }
        drop(txo);
        *self.done.lock().unwrap() = true;
        self.done_cv.notify_all();
    }
}

fn make_size_string(u: u64) -> String {
    //format!("{} GB", commas(u / 1_000_000_000))
    format!("{} GB", u as f64 / 1_000_000_000.0)
}

fn commas(n: u64) -> String {
    let num_str = n.to_string();
    let mut s = String::new();
    // Track when to place comma with cc
    let mut cc = -1;
    for num in num_str.chars().rev() {
        cc += 1;
        if cc == 3 {
            s = format!("{},{}", num, s);
            cc = 0;
        } else {
            s = format!("{}{}", num, s);
        }
    }
    s
}

fn parse_args() -> Args {
    let mut args = Args::default();
    let mut flag = None;
    for arg in std::env::args().skip(1) {
        if arg.starts_with("-") {
            if let Some(name) = flag.take() {
                die!(1, "Expected value for {name}");
            }
            flag = match_flag(arg, &mut args);
        } else if let Some(name) = flag.take() {
            match_flag(format!("{name}={arg}"), &mut args);
        } else {
            args.paths.push(arg.into());
        }
    }
    args
}

fn match_flag(arg: String, args: &mut Args) -> Option<String> {
    let (name, val, have_val) = if let Some((name, val)) = arg.split_once('=') {
        (name, val, true)
    } else {
        (arg.as_str(), "", false)
    };
    match name {
        "--sort" => {
            if have_val {
                die!(1, "{name} doesn't expect value");
            }
            args.sort = true;
        }
        "-t" | "--threads" => {
            if !have_val {
                return Some(arg);
            }
            let Ok(val) = val.parse::<usize>() else {
                die!(1, "Invalid value for {name}, expected non-negative integer")
            };
            args.threads = val.max(2);
        }
        "--help" | "-h" => {
            if have_val {
                eprintln!("{name} doesn't expect value");
            }
            let args = Args::default();
            die!(0, "\
Usage: finfo [FLAGS] [FILES...]
    --sort\n\tSort the results
    -t, --threads\n\tNumber of worker threads running (total number of threads will be 1 + this value) (default: {}, min: 1)
",
                args.threads,
            );
        }
        _ => die!(1, "Unknown flag: {arg}"),
    }
    None
}
