use clap::Parser;

#[derive(Parser)]
pub struct Config {
    #[arg(long, env = "CLIP_HOST", default_value = "0.0.0.0")]
    pub host: String,
    #[arg(long, env = "CLIP_PORT", default_value = "50051")]
    pub port: u16,
    #[arg(long, env = "CLIP_MODEL_NAME")]
    pub model_name: String,
    #[arg(long, env = "CLIP_MODEL_PATH")]
    pub model_path: String,
    #[arg(long, env = "CLIP_TOKENIZER_PATH")]
    pub tokenizer_path: String,
    #[arg(long, env = "CLIP_CPU_ONLY", default_value = "false")]
    pub cpu: bool,
    #[arg(long, env = "CLIP_CONCURRENCY", default_value = "1")]
    pub concurrency: u32,
} 
