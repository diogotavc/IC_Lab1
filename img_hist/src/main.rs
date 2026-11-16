use std::collections::HashMap;
use std::error::Error;
use std::hash::Hash;
use std::{env, ffi::OsStr, path::Path};
use std::io::Cursor;
use image::{ImageReader, Luma, Pixel, Rgb};
use std::fs::File;
use std::io::Write;
use image::GenericImageView;

fn main() -> Result<(), Box<dyn Error>> {
    let args: Vec<String> = env::args().collect();
    // dbg!(&args[1]);

    // if Path::new(&args[1]).extension() == Some(OsStr::new("raw")) {
        
    // }
    let img = ImageReader::open(&args[1])?.decode()?;
    //let mut file = File::create("output.txt")?;
    //writeln!(&mut file, "{:?}", img)?;
    //println!("dimensions: {:?}", img.dimensions());
    //println!("color: {:?}", img.color());
    // println!("{:?}", img);
    match img {
        image::DynamicImage::ImageLuma8(image_buffer) => {
            let mut counts = [0_usize; 256];
            for &Luma([value]) in image_buffer.pixels() {
                counts[value as usize] += 1;
            }
            let mut file = File::create("lumaChannel8.txt")?;
            writeln!(&mut file, "{:?}", counts)?;
        },
        image::DynamicImage::ImageLuma16(image_buffer) => {
            let mut counts = [0_usize; 0x10000];
            for &Luma([value]) in image_buffer.pixels() {
                counts[value as usize] += 1;
            }
            let mut file = File::create("lumaChannel16.txt")?;
            writeln!(&mut file, "{:?}", counts)?;
        },
        image::DynamicImage::ImageRgb8(image_buffer) => {
            let mut countsR = [0_usize; 256];
            let mut countsG = [0_usize; 256];
            let mut countsB = [0_usize; 256];
            for &Rgb([R,G,B]) in image_buffer.pixels() {
                countsR[R as usize] += 1;
                countsG[G as usize] += 1;
                countsB[B as usize] += 1;
            }
            let mut file = File::create("RChannel8.txt")?;
            writeln!(&mut file, "{:?}", countsR)?;
            file = File::create("GChannel8.txt")?;
            writeln!(&mut file, "{:?}", countsG)?;
            file = File::create("BChannel8.txt")?;
            writeln!(&mut file, "{:?}", countsB)?;
        },
        image::DynamicImage::ImageRgb16(image_buffer) => {
            let mut countsR = [0_usize; 0x10000];
            let mut countsG = [0_usize; 0x10000];
            let mut countsB = [0_usize; 0x10000];
            for &Rgb([R,G,B]) in image_buffer.pixels() {
                countsR[R as usize] += 1;
                countsG[G as usize] += 1;
                countsB[B as usize] += 1;
            }
            let mut file = File::create("RChannel16.txt")?;
            writeln!(&mut file, "{:?}", countsR)?;
            file = File::create("GChannel16.txt")?;
            writeln!(&mut file, "{:?}", countsG)?;
            file = File::create("BChannel16.txt")?;
            writeln!(&mut file, "{:?}", countsB)?;
        },
        _ => panic!("unsupported image format"),
    }

    Ok(())
}

