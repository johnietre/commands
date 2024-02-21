// TODO: replace file names
// TODO: Count occurrences of file name matches
// TODO: Handle InvalidData errors
use clap::Parser;
use crossbeam::channel::{unbounded, Receiver, Sender};
use regex::Regex;
use std::fs;
use std::io::{prelude::*, BufReader, BufWriter, ErrorKind};
use std::path::{Path, PathBuf};
use std::process::exit;
use std::sync::Arc;
use std::thread;

type Res<T> = Result<T, Box<dyn std::error::Error>>;

fn main() {
    let args = Args::parse();
    let app = match App::new(args) {
        Ok(app) => app,
        Err(e) => {
            eprintln!("{e}");
            exit(1);
        }
    };
    app.run();
}

struct WorkerData(PathBuf, Sender<WorkerData>);

struct App {
    args: Args,
    what: What,
}

impl App {
    fn new(mut args: Args) -> Res<Arc<Self>> {
        let mut what = std::mem::replace(&mut args.what, String::new());
        if what.starts_with('\\') {
            let mut remove = false;
            for c in what.chars().skip(1) {
                if c == '-' {
                    remove = true;
                } else if c != '\\' {
                    break;
                }
            }
            if remove {
                what = what.split_off(1);
            }
        }
        if args.insensitive {
            what.make_ascii_lowercase();
        }
        let what = if args.regex {
            What::Regex(Regex::new(&what)?)
        } else {
            What::Text(what)
        };
        if args.paths.len() == 0 {
            args.paths = vec![PathBuf::from(".")];
        }
        Ok(Arc::new(Self { args, what }))
    }

    fn run(self: Arc<Self>) {
        let (results_tx, results_rx) = unbounded();
        let (worker_tx, worker_rx) = unbounded();
        let threads = (1..=self.args.threads.max(1))
            .filter_map(|i| {
                let app = Arc::clone(&self);
                let (worker_rx, results_tx) = (worker_rx.clone(), results_tx.clone());
                let name = format!("#{}", i);
                let handle = thread::Builder::new()
                    .name(name.clone())
                    .spawn(move || app.run_worker(worker_rx, results_tx));
                match handle {
                    Ok(handle) => Some(handle),
                    Err(e) => {
                        eprintln!("error spawning thread {name}: {e}");
                        None
                    }
                }
            })
            .collect::<Vec<_>>();

        drop(results_tx);
        for path in &self.args.paths {
            if worker_tx
                .send(WorkerData(path.clone(), worker_tx.clone()))
                .is_err()
            {
                eprintln!("fatal error: all worker receivers dropped");
                exit(1);
            }
        }
        drop(worker_tx);
        drop(worker_rx);

        let mut results = Vec::new();
        for found in results_rx {
            let found = found.strip_prefix("./").unwrap_or(&found).to_string();
            if self.args.sort {
                results.push(found);
            } else {
                println!("{found}");
            }
        }
        if results.len() != 0 {
            results.sort();
            for res in results {
                println!("{res}");
            }
        }

        threads.into_iter().for_each(|h| {
            let thread = h.thread();
            let (name, id) = (thread.name().map(|s| s.to_string()), thread.id());
            if h.join().is_err() {
                if let Some(name) = name {
                    eprintln!("error joining thread {name}");
                } else {
                    eprintln!("error joining thread (id: {id:?})");
                }
            }
        });
    }

    fn run_worker(self: &Arc<Self>, worker_rx: Receiver<WorkerData>, results_tx: Sender<String>) {
        for WorkerData(path, worker_tx) in worker_rx {
            let md = match fs::metadata(&path) {
                Ok(md) => md,
                Err(e) => {
                    if !self.args.mute {
                        eprintln!("{}: error getting metadata: {e}", path.display());
                    }
                    continue;
                }
            };
            if md.is_dir() {
                self.search_dir(path, worker_tx, &results_tx);
            } else if md.is_file() {
                self.search_file(path, &results_tx);
            }
        }
    }

    fn search_dir(
        &self,
        path: impl AsRef<Path>,
        worker_tx: Sender<WorkerData>,
        results_tx: &Sender<String>,
    ) {
        let path = path.as_ref();
        let dir_iter = match path.read_dir() {
            Ok(iter) => iter,
            Err(e) => {
                if !self.args.mute {
                    eprintln!("{}: error reading directory: {e}", path.display());
                }
                return;
            }
        };
        // TODO: Too many files open
        for ent in dir_iter {
            let ent = match ent {
                Ok(ent) => ent,
                Err(e) => {
                    if !self.args.mute {
                        eprintln!("{}: error getting directory entry: {e}", path.display());
                    }
                    continue;
                }
            };
            let ftype = match ent.file_type() {
                Ok(ftype) => ftype,
                Err(e) => {
                    if !self.args.mute {
                        eprintln!("{}: error getting file type: {e}", path.display());
                    }
                    continue;
                }
            };
            if ftype.is_symlink() {
                continue;
            }
            let fname = ent.file_name();
            let Some(name) = fname.to_str() else {
                // TODO
                continue;
            };
            let full_path = path.join(name);
            if self.args.content && !ftype.is_dir() {
                self.search_file(full_path, results_tx);
                continue;
            } else if !self.args.content && self.args.replace.is_none() {
                if self.what.matches(name) {
                    let output = if ftype.is_dir() {
                        format!("{}/", full_path.display())
                    } else {
                        format!("{}", full_path.display())
                    };
                    if results_tx.send(output).is_err() {
                        eprintln!("fatal error: results receiver has been dropped");
                        exit(1);
                    }
                }
            }
            if self.args.recursive && ftype.is_dir() {
                if worker_tx
                    .send(WorkerData(full_path, worker_tx.clone()))
                    .is_err()
                {
                    eprintln!("fatal error: worker receiver has been dropped");
                    exit(1);
                }
            }
        }
    }

    fn search_file(&self, path: impl AsRef<Path>, results_tx: &Sender<String>) {
        let path = path.as_ref();
        if let Some(ext) = path.extension() {
            if self
                .args
                .ignore_exts
                .iter()
                .position(|s| s.as_os_str() == ext)
                .is_some()
            {
                return;
            }
        }
        match infer::get_from_path(&path) {
            Ok(Some(kind)) => {
                if kind.matcher_type() != infer::MatcherType::Text {
                    return;
                }
            }
            Ok(None) => (),
            Err(e) => {
                if !self.args.mute {
                    eprintln!("{}: error inferring file type: {e}", path.display());
                }
                return;
            }
        }
        let file = match fs::File::open(&path) {
            Ok(file) => file,
            Err(e) => {
                if !self.args.mute {
                    eprintln!("{}: error opening file: {e}", path.display());
                }
                return;
            }
        };
        // TODO: Handle too many files
        if let Some(to) = self.args.replace.as_ref() {
            self.replace_file_content(path, file, to);
            return;
        }
        self.search_file_content(path, file, results_tx);
    }

    fn search_file_content(
        &self,
        path: impl AsRef<Path>,
        file: fs::File,
        results_tx: &Sender<String>,
    ) {
        let path = path.as_ref();
        let (mut count, mut linenos) = (0, Vec::new());
        for (i, line) in BufReader::new(file).lines().enumerate() {
            let mut line = match line {
                Ok(line) => line,
                Err(e) if e.kind() == ErrorKind::InvalidData => break,
                Err(e) => {
                    if !self.args.mute {
                        eprintln!("{}: error reading line: {e}", path.display());
                    }
                    return;
                }
            };
            if self.args.insensitive {
                line.make_ascii_lowercase();
            }
            if self.what.matches(&line) {
                if self.args.linenos {
                    linenos.push(i + 1);
                } else if self.args.counts {
                    count += 1;
                } else {
                    if results_tx.send(path.display().to_string()).is_err() {
                        eprintln!("fatal error: results receiver has been dropped");
                        exit(1);
                    }
                    return;
                }
            }
        }
        let output = if self.args.linenos {
            if linenos.len() == 0 {
                return;
            }
            count = linenos.len();
            let linenos = linenos
                .into_iter()
                .map(|l| l.to_string())
                .collect::<Vec<_>>()
                .join(",");
            if self.args.counts {
                format!("{} | {linenos} | ({count})", path.display())
            } else {
                format!("{} | {linenos}", path.display())
            }
        } else if self.args.counts && count != 0 {
            format!("{} | ({count})", path.display())
        } else {
            return;
        };
        if results_tx.send(output).is_err() {
            eprintln!("fatal error: results receiver has been dropped");
            exit(1);
        }
    }

    fn replace_file_content(&self, path: impl AsRef<Path>, file: fs::File, to: &str) {
        let path = path.as_ref();
        let Some(fname) = path.file_name().map(|s| s.to_str()).flatten() else {
            return;
        };
        let temp_path = path.with_file_name(format!("{}.jtsearch", fname));
        let mut writer = match fs::File::create(&temp_path) {
            Ok(f) => BufWriter::new(f),
            Err(e) => {
                if !self.args.mute {
                    eprintln!("{}: error creating temp file: {e}", temp_path.display());
                }
                return;
            }
        };
        for line in BufReader::new(file).lines() {
            let line = match line {
                Ok(line) => line,
                Err(e) if e.kind() == ErrorKind::InvalidData => break,
                Err(e) => {
                    if !self.args.mute {
                        eprintln!("{}: error reading line: {e}", path.display());
                    }
                    if let Err(e) = fs::remove_file(&temp_path) {
                        if !self.args.mute {
                            eprintln!("{}: error removing file: {e}", temp_path.display());
                        }
                    }
                    return;
                }
            };
            // TODO: Handle case (i.e., insensitive)
            let line = self.what.replace(&line, to);
            if let Err(e) = write!(writer, "{line}\n") {
                if !self.args.mute {
                    eprintln!("{}: error writing to file: {e}", temp_path.display());
                }
                if let Err(e) = fs::remove_file(&temp_path) {
                    if !self.args.mute {
                        eprintln!("{}: error removing file: {e}", temp_path.display());
                    }
                }
                return;
            }
        }
        if let Err(e) = writer.flush() {
            if !self.args.mute {
                eprintln!("{}: error flushing to file: {e}", temp_path.display());
            }
            if let Err(e) = fs::remove_file(&temp_path) {
                if !self.args.mute {
                    eprintln!("{}: error removing file: {e}", temp_path.display());
                }
            }
            return;
        }
        if let Err(e) = fs::rename(&temp_path, path) {
            if !self.args.mute {
                eprintln!(
                    "{}, {}: error renaming file: {e}",
                    temp_path.display(),
                    path.display()
                );
            }
            return;
        }
    }

    #[allow(dead_code)]
    fn print_err(&self, e: impl std::fmt::Display) {
        if !self.args.mute {
            eprintln!("{e}");
        }
    }
}

enum What {
    Text(String),
    Regex(Regex),
}

impl What {
    fn matches(&self, text: &str) -> bool {
        match self {
            What::Text(s) => text.contains(s),
            What::Regex(r) => r.is_match(text),
        }
    }

    fn replace(&self, text: &str, to: &str) -> String {
        match self {
            What::Text(s) => text.replace(s, to),
            What::Regex(r) => r.replace_all(text, to).into_owned(),
        }
    }
}

/// A program to find stuff. Does not follow symbolic links.
#[derive(Parser, Debug)]
#[command(about, long_about = None)]
struct Args {
    /// What to search for. Prefix with '\' if it starts with a '-' and yoou don't want it to be
    /// treated as a flag. Only works when '-' is only preceded by '\'s (i.e., "\--help" becomes
    /// "--help" and "\\--help" becomes "\--help" internally)
    what: String,

    /// Search file contents
    #[arg(short, long)]
    content: bool,

    /// Search directories recursively
    #[arg(short, long)]
    recursive: bool,

    /// Case-insensitive search
    #[arg(short, long)]
    insensitive: bool,

    /// Sort results
    #[arg(short, long)]
    sort: bool,

    /// Mute non-fatal errors
    #[arg(short, long)]
    mute: bool,

    /// Count the number of occurences
    #[arg(short = 'n', long)]
    counts: bool,

    /// Use regex
    #[arg(short = 'x', long)]
    regex: bool,

    /// Print line numbers of occurrences
    #[arg(short, long)]
    linenos: bool,

    /// Text to replace matches with
    #[arg(short = 'p', long)]
    replace: Option<String>,

    /// Ignored file types
    #[arg(long)]
    ignore_exts: Vec<std::ffi::OsString>,

    /*
    /// Ignored paths
    #[arg]
    ignore_paths: Vec<String>,
    */
    /// Number of threads to use
    #[arg(short, long, default_value_t = num_cpus::get())]
    threads: usize,

    /// Files/directories to search
    paths: Vec<PathBuf>,
}
