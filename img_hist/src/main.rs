use std::error::Error;
use std::{env};
use image::{ImageReader, Luma, Rgb};
use std::fs::File;
use std::io::Write;

fn calculate_entropy(counts: &[usize]) -> f64 {
    let sum = counts.iter().sum::<usize>() as f64;
    let mut h = 0.0;

    for &c in counts {
        if c == 0 {
            continue;
        }
        let p = c as f64 / sum;
        h -= p * p.log2();
    }

    h
}

fn main() -> Result<(), Box<dyn Error>> {
    let args: Vec<String> = env::args().collect();
    let img = ImageReader::open(&args[1])?.decode()?;

    match img {
        image::DynamicImage::ImageLuma8(image_buffer) => {
            let mut counts = [0_usize; 256];
            for &Luma([value]) in image_buffer.pixels() {
                counts[value as usize] += 1;
            }
            let mut file = File::create("lumaChannel8.txt")?;
            writeln!(&mut file, "{:?}", counts)?;
            println!("{:?}", calculate_entropy(&counts));
        },
        image::DynamicImage::ImageLuma16(image_buffer) => {
            let mut counts = [0_usize; 0x10000];
            for &Luma([value]) in image_buffer.pixels() {
                counts[value as usize] += 1;
            }
            let mut file = File::create("lumaChannel16.txt")?;
            writeln!(&mut file, "{:?}", counts)?;
            println!("{:?}", calculate_entropy(&counts));
        },
        image::DynamicImage::ImageRgb8(image_buffer) => {
            let mut counts_r = [0_usize; 256];
            let mut counts_g = [0_usize; 256];
            let mut counts_b = [0_usize; 256];
            for &Rgb([r,g,b]) in image_buffer.pixels() {
                counts_r[r as usize] += 1;
                counts_g[g as usize] += 1;
                counts_b[b as usize] += 1;
            }
            let mut file = File::create("RChannel8.txt")?;
            writeln!(&mut file, "{:?}", counts_r)?;
            println!("{:?}", calculate_entropy(&counts_r));
            file = File::create("GChannel8.txt")?;
            writeln!(&mut file, "{:?}", counts_g)?;
            println!("{:?}", calculate_entropy(&counts_g));
            file = File::create("BChannel8.txt")?;
            writeln!(&mut file, "{:?}", counts_b)?;
            println!("{:?}", calculate_entropy(&counts_b));
        },
        image::DynamicImage::ImageRgb16(image_buffer) => {
            let mut counts_r = [0_usize; 0x10000];
            let mut counts_g = [0_usize; 0x10000];
            let mut counts_b = [0_usize; 0x10000];
            for &Rgb([r,g,b]) in image_buffer.pixels() {
                counts_r[r as usize] += 1;
                counts_g[g as usize] += 1;
                counts_b[b as usize] += 1;
            }
            let mut file = File::create("RChannel16.txt")?;
            writeln!(&mut file, "{:?}", counts_r)?;
            println!("{:?}", calculate_entropy(&counts_r));
            file = File::create("GChannel16.txt")?;
            writeln!(&mut file, "{:?}", counts_g)?;
            println!("{:?}", calculate_entropy(&counts_g));
            file = File::create("BChannel16.txt")?;
            writeln!(&mut file, "{:?}", counts_b)?;
            println!("{:?}", calculate_entropy(&counts_b));
        },
        _ => panic!("unsupported image format"),
    }

    Ok(())
}
