CREATE TABLE IF NOT EXISTS expenses (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    date date NOT NULL,
    spent_at text NOT NULL,
    notes text,
    category text NOT NULL,
    payment_method text NOT NULL,
    iso_currency_code text NOT NULL,
    amount numeric(8,2) NOT NULL CHECK (amount BETWEEN 0 AND 1000000),
    version integer NOT NULL DEFAULT 1
);