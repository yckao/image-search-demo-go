use futures::StreamExt;
use tonic::{Request, Response, Status, Streaming};

use candle_core::Tensor;
use candle_transformers::models::clip;

use crate::model::tokenize;
use crate::proto::clip::{
    clip_service_server::ClipService, EmbeddingResponse, ImageChunk, Text,
};
use crate::utils::load_image;

use std::sync::Arc;
use tokio::sync::Semaphore;

pub struct ClipServiceImpl {
    pub model_name: String,
    pub model: clip::ClipModel,
    pub config: clip::ClipConfig,
    pub tokenizer: tokenizers::Tokenizer,
    pub device: candle_core::Device,
    pub semaphore: Arc<Semaphore>,
}

#[tonic::async_trait]
impl ClipService for ClipServiceImpl {
    async fn image_embedding(
        &self,
        request: Request<Streaming<ImageChunk>>,
    ) -> Result<Response<EmbeddingResponse>, Status> {
        let _permit = self.semaphore.acquire().await.map_err(|_e| {
            Status::internal("Failed to acquire GPU semaphore")
        })?;

        let mut stream = request.into_inner();
        let mut image_data: Vec<u8> = Vec::new();

        while let Some(chunk) = stream.next().await {
            image_data.extend_from_slice(&chunk?.data);
        }

        let img = match load_image(&image_data, self.config.image_size) {
            Ok(img) => img,
            Err(e) => return Err(Status::invalid_argument(e.to_string())),
        }.to_device(&self.device).map_err(|e| Status::internal(e.to_string()))?;

        let embedding = clip::div_l2_norm(
            &self
                .model
                .get_image_features(
                    &Tensor::stack(&[img], 0).map_err(|e| Status::internal(e.to_string()))?
                )
                .map_err(|e| Status::internal(e.to_string()))?
                .flatten_all()
                .map_err(|e| Status::internal(e.to_string()))?,
        )
        .map_err(|e| Status::internal(e.to_string()))?
        .to_vec1::<f32>()
        .map_err(|e| Status::internal(e.to_string()))?;

        Ok(Response::new(EmbeddingResponse {
            model_name: self.model_name.clone(),
            embedding,
        }))
    }

    async fn text_embedding(
        &self,
        request: Request<Text>,
    ) -> Result<Response<EmbeddingResponse>, Status> {
        let _permit = self.semaphore.acquire().await.map_err(|_e| {
            Status::internal("Failed to acquire GPU semaphore")
        })?;

        let text = request.into_inner().text;
        let input_ids = tokenize(text, &self.tokenizer, &self.device)
            .map_err(|e| Status::internal(e.to_string()))?;
        let embedding = clip::div_l2_norm(
            &self
                .model
                .get_text_features(&input_ids)
                .map_err(|e| Status::internal(e.to_string()))?
                .flatten_all()
                .map_err(|e| Status::internal(e.to_string()))?,
        )
        .map_err(|e| Status::internal(e.to_string()))?
        .to_vec1::<f32>()
        .map_err(|e| Status::internal(e.to_string()))?;

        Ok(Response::new(EmbeddingResponse {
            model_name: self.model_name.clone(),
            embedding,
        }))
    }
} 