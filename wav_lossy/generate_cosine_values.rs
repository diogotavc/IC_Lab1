use std::{f64::consts::PI, fs::File, error::Error, io::{prelude::*, BufWriter}};

fn main() -> Result<(), Box<dyn Error>> {
    const CHUNK_SIZE: usize = 8192;
    
    let mut output_file = BufWriter::new(File::create("cosine_values.bin")?);
    for f in 0..CHUNK_SIZE {
        for i in 0..CHUNK_SIZE {
            output_file.write_all(&f64::cos(f as f64 * (i as f64 + 0.5) * PI / CHUNK_SIZE as f64).to_le_bytes())?;
        }
    }

    Ok(())
}
