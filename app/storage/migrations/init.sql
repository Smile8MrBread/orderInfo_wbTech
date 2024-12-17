CREATE TABLE "Delivery"
(
    id      BIGSERIAL PRIMARY KEY,
    name    VARCHAR(255),
    phone   VARCHAR(15),
    zip     VARCHAR(7),
    city    VARCHAR(255),
    address VARCHAR(255),
    region  VARCHAR(255),
    email   VARCHAR(255)
);

CREATE TABLE "Payment"
(
    id            BIGSERIAL PRIMARY KEY,
    transaction   VARCHAR(255) ,
    request_id    VARCHAR(255),
    currency      VARCHAR(7),
    provider      VARCHAR(255),
    amount        DECIMAL(10, 2),
    payment_dt    TIMESTAMP,
    bank          VARCHAR(255),
    delivery_cost DECIMAL(10, 2),
    goods_total   DECIMAL(10, 2),
    custom_fee    INT
);

CREATE TABLE "Order"
(
    order_uid          VARCHAR(255) PRIMARY KEY,
    track_number       VARCHAR(255) ,
    entry              VARCHAR(255),
    delivery_id        INT REFERENCES "Delivery",
    payment_id         INT REFERENCES "Payment",
    locale             VARCHAR(3),
    internal_signature VARCHAR(255),
    customer_id        VARCHAR(255),
    delivery_service   VARCHAR(255),
    shardkey           VARCHAR(255),
    sm_id              INT,
    date_created       TIMESTAMP,
    oof_shard          VARCHAR(255)
);

CREATE TABLE "Items"
(
    id           BIGSERIAL PRIMARY KEY,
    order_id     VARCHAR(255) REFERENCES "Order",
    chrt_id      INT,
    track_number VARCHAR(255) ,
    price        DECIMAL(10, 2),
    rid          VARCHAR(255),
    name         VARCHAR(255),
    sale         DECIMAL(10, 2),
    size         VARCHAR(127),
    total_price  DECIMAL(10, 2),
    nm_id        INT,
    brand        VARCHAR(255),
    status       INT
);
