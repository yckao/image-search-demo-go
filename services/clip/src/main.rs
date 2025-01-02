use clap::Parser;
use tonic::transport::Server;

use candle_nn::VarBuilder;
use candle_transformers::models::clip as clip_model;

mod config;
mod model;
mod proto;
mod service;
mod utils;

use config::Config;
use proto::clip::clip_service_server::ClipServiceServer;
use service::ClipServiceImpl;
use utils::device;

use std::sync::Arc;
use tokio::sync::Semaphore;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Config::parse();
    let addr = format!("{}:{}", args.host, args.port).parse().unwrap();

    let device = device(args.cpu)?;
    let config = clip_model::ClipConfig::vit_base_patch32();
    let vb = unsafe {
        VarBuilder::from_mmaped_safetensors(&[&args.model_path], candle_core::DType::F32, &device)?
    };
    let model: clip_model::ClipModel = clip_model::ClipModel::new(vb, &config)?;
    let tokenizer = tokenizers::Tokenizer::from_file(&args.tokenizer_path)
        .map_err(anyhow::Error::msg)?;
    let service = ClipServiceImpl {
        model_name: args.model_name,
        model,
        config,
        tokenizer,
        device,
        semaphore: Arc::new(Semaphore::new(args.concurrency as usize)),
    };

    println!("Server listening on {}", addr);

    Server::builder()
        .add_service(ClipServiceServer::new(service))
        .serve(addr)
        .await?;
    Ok(())
}
