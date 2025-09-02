use sqlx::{Pool, Postgres};

pub async fn init_db() -> Result<Pool<Postgres>, sqlx::Error> {
    let database_url = std::env::var("DATABASE_URL").expect("DATABASE_URL must be set in .env");

    let pool = Pool::<Postgres>::connect(&database_url).await?;
    Ok(pool)
}
