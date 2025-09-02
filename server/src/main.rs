mod model;
mod router;

use actix_web::{App, HttpServer};
use model::db::init_db;

#[tokio::main]
async fn main() -> std::io::Result<()> {
    dotenvy::dotenv().ok();

    // initialize db connection pool
    let pool = init_db().await.expect("Failed to connect to DB");

    HttpServer::new(move || {
        App::new()
            .app_data(pool.clone()) // share pool with routes
            .configure(router::init)
    })
    .bind(("127.0.0.1", 8080))?
    .run()
    .await
}
