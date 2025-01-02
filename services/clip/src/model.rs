use anyhow::Error as E;
use candle_core::{Device, Tensor};
use tokenizers::Tokenizer;

pub fn tokenize(
    sequence: String,
    tokenizer: &Tokenizer,
    device: &Device,
) -> anyhow::Result<Tensor> {
    let token = tokenizer
        .encode(sequence, true)
        .map_err(E::msg)?
        .get_ids()
        .to_vec();
    let input_ids = Tensor::new(vec![token], device)?;

    Ok(input_ids)
} 