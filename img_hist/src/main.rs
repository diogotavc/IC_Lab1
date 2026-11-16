//use std::collections::HashMap;
use std::error::Error;
//use std::hash::Hash;
// use std::{env, ffi::OsStr, path::Path};
use std::{array, env};
//use std::io::Cursor;
use image::{DynamicImage, ImageReader, Luma, Rgb};
use std::fs::File;
use std::io::Write;
//use image::GenericImageView;

fn entropy() -> Result<u32, Box<dyn Error>> {
    Ok(0)
}

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
            file = File::create("GChannel8.txt")?;
            writeln!(&mut file, "{:?}", counts_g)?;
            file = File::create("BChannel8.txt")?;
            writeln!(&mut file, "{:?}", counts_b)?;
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
            file = File::create("GChannel16.txt")?;
            writeln!(&mut file, "{:?}", counts_g)?;
            file = File::create("BChannel16.txt")?;
            writeln!(&mut file, "{:?}", counts_b)?;
        },
        _ => panic!("unsupported image format"),
    }

    Ok(())
}

