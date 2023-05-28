use std::fs;
use std::io;
use std::path::Path;
use std::sync::mpsc::{sync_channel, Receiver};
use std::thread;
use std::time;

pub fn watch(path: impl AsRef<Path>) -> io::Result<Receiver<FileEvent>> {
    let path = path.as_ref().to_path_buf();
    let (tx, rx) = sync_channel(1);
    let mut prev_info = fs::metadata(&path)?;
    // Call to make sure there's no error getting the mod time.
    prev_info.modified()?;
    thread::spawn(move || {
        loop {
            let _ = tx.send(FileEvent::Write);
            thread::sleep(time::Duration::from_millis(500));
            let info = match fs::metadata(&path) {
                Ok(info) => info,
                Err(e) => {
                    if e.kind() == io::ErrorKind::NotFound {
                        let _ = tx.send(FileEvent::Remove);
                    } else {
                        let _ = tx.send(FileEvent::Error(e));
                    }
                    return;
                }
            };
            if info.len() != prev_info.len() {
                let _ = tx.send(FileEvent::Write);
            } else {
                match (info.modified(), prev_info.modified()) {
                    (Ok(t), Ok(pt)) if t != pt => {
                        let _ = tx.send(FileEvent::Write);
                    }
                    (Err(e), _) => {
                        // Only the newest info needs to be checked for error since the previous
                        // info shouldn't produce a new error on a repeat call.
                        let _ = tx.send(FileEvent::Error(e));
                        return;
                    }
                    _ => (),
                }
            }
            prev_info = info;
        }
    });
    Ok(rx)
}

pub enum FileEvent {
    Write,
    Remove,
    Error(io::Error),
}
