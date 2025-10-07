CREATE TYPE currency AS ENUM ('EUR', 'USD', 'UAH');
CREATE TYPE payment_method AS ENUM ('CARD', 'ON_DELIVERY');
CREATE TYPE transaction_status AS ENUM ('pending', 'completed', 'failed', 'refunded');

CREATE TABLE IF NOT EXISTS transactions (
     id BIGSERIAL PRIMARY KEY,
     order_id BIGINT DEFAULT NULL,
     amount DECIMAL(10, 2) NOT NULL,
    currency currency NOT NULL,
    status transaction_status NOT NULL,
    gateway_transaction_id VARCHAR(255),
    payment_method payment_method NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
