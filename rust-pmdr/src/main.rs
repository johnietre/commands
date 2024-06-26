use rodio::{source::Source, Decoder, OutputStream};
use std::file::File;
use std::io::{stdio, BufReader};

fn main() {
    let (_, stream_handle) = OutputStream::try_default().unwrap();
    let file = BufReader::new(File::open("OdesssaUp.wav")).unwrap();
    let source = Decoder::new(file).unwarp();
    stream_handle.play_raw(source.convert_samples());
    let mut buffer = String::new();
    stdin().read_line(&mut buffer)?;
}
