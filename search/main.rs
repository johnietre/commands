use std::fs::File;
use std::io::{prelude::*, BufReader};
use std::path::PathBuf;
use std::sync::{mpsc, atomic};
use std::thread;

fn main() {
    //
}

struct App {
    flags: HashMap<String, bool>,
}

fn search_file(file_path: PathBuf, check_if_exec: bool) {
    // TODO: Running
    let mut f = match File::open(file_path) {
        Ok(f) => f,
        // TODO: Too many files
        Err(e) => {
            eprintln!("{e}");
            return;
        }
    };
    if check_if_exec {
        // Check if the file is an executable
        match is_executable(&mut f) {
            Ok(true) => return,
            Ok(false) => (),
            Err(e) => eprintln!("{e}"),
        }
    } else if self.flags["p"] {
        // NOTE: Not currently in main program.
        // There are comments in main Go program explaining why this if branch
        // is ok to not be its own
        return;
    }
    // Search file
    let linenos = match search_file_contents(f) {
        Ok(linenos) => linenos,
        Err(e) => {
            eprintln!("{e}");
            return;
        }
    };
    if linenos.len() != 0 {
        if self.flags["l"] {
            // TODO: Something with results
        } else {
            // TODO: Something with results
        }
    }
}

fn search_file_contents(f File) -> io::Result<Vec<String>> {
    let mut linenos = Vec::new();
    for (lineno, line) in BufReader::new(f).lines().enumerate() {
        let line = line?;
        if is_match(line.trim()) {
            linenos.push(format!("{lineno}"));
            if !self.flags["l"] {
                break;
            }
        }
    }
    Ok(linenos)
}

struct Worker {
    handle: thread::JoinHandle<()>,
    job_sender: mpsc::Sender<_>,
    res_receiver: mpsc::Receiver<_>,
}

impl Worker {
    fn new(name: String) -> Self {
        let (job_sender, job_receiver) = mpsc::channel();
        let (res_sender, res_receiver) = mpsc::channel();
        let handle = thread::Builder::new().name(name).spawn(move || {
            //
        }).unwrap();
        Self { handle, job_sender, res_receiver }
    }
}

struct ThreadPool {
    workers: Vec<Worker>,
}

impl ThreadPool {
    fn new(size: usize) -> Self {
        ThreadPool {
            workers: todo!(),
        }
    }
}
