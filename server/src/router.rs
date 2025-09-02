use actix_web::{HttpResponse, Responder, get, web};

#[get("/healthCheck")]
async fn health_check() -> impl Responder {
    HttpResponse::Ok().body("Server is working!")
}

pub fn init(cfg: &mut web::ServiceConfig) {
    cfg.service(web::scope("/api/v1").service(health_check));
}
