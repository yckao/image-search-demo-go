use std::io::Cursor;

use candle_core::{DType, Device, Result, Tensor};

pub fn load_image(image_data: &[u8], image_size: usize) -> anyhow::Result<Tensor> {
    let img = image::ImageReader::new(Cursor::new(image_data))
        .with_guessed_format()?
        .decode()?;

    let img = img.to_rgb8();
    let (width, height) = img.dimensions();
    let (new_width, new_height) = if width < height {
        let new_width = image_size as u32;
        let new_height = (height as f32 * (image_size as f32 / width as f32)).round() as u32;
        (new_width, new_height)
    } else {
        let new_height = image_size as u32;
        let new_width = (width as f32 * (image_size as f32 / height as f32)).round() as u32;
        (new_width, new_height)
    };
    let resized: image::ImageBuffer<_, Vec<_>> = image::imageops::resize(
        &img,
        new_width,
        new_height,
        image::imageops::FilterType::CatmullRom, // or another resampling method
    );

    let (resized_w, resized_h) = resized.dimensions();
    let x = (resized_w - image_size as u32) / 2;
    let y = (resized_h - image_size as u32) / 2;
    let cropped = image::imageops::crop_imm(&resized, x, y, image_size as u32, image_size as u32).to_image();

    let rescale_factor = 1.0 / 255.0;

    let image_mean = [0.48145466, 0.4578275, 0.40821073];
    let image_std = [0.26862954, 0.26130258, 0.27577711];

    let mut output = Vec::with_capacity(image_size * image_size * 3);
    for pixel in cropped.pixels() {
        let [r, g, b] = pixel.0;
        for (i, &val) in [r, g, b].iter().enumerate() {
            let x = val as f32 * rescale_factor;
            let x = (x - image_mean[i]) / image_std[i];
            output.push(x);
        }
    }

    let img = Tensor::from_vec(output, (224, 224, 3), &Device::Cpu)?
        .permute((2, 0, 1))?
        .to_dtype(DType::F32)?;

    Ok(img)
}

pub fn device(cpu: bool) -> Result<Device> {
    if cpu {
        Ok(Device::Cpu)
    } else if candle_core::utils::cuda_is_available() {
        Ok(Device::new_cuda(0)?)
    } else if candle_core::utils::metal_is_available() {
        Ok(Device::new_metal(0)?)
    } else {
        #[cfg(all(target_os = "macos", target_arch = "aarch64"))]
        {
            println!(
                "Running on CPU, to run on GPU(metal), build this example with `--features metal`"
            );
        }
        #[cfg(not(all(target_os = "macos", target_arch = "aarch64")))]
        {
            println!("Running on CPU, to run on GPU, build this example with `--features cuda`");
        }
        Ok(Device::Cpu)
    }
} 