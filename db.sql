CREATE TABLE user_payments (
                                             id SERIAL PRIMARY KEY,
                                             user_id VARCHAR(255) UNIQUE NOT NULL,
                                             invoice_id VARCHAR(255) NOT NULL,
                                             status VARCHAR(50) DEFAULT 'pending',
                                             amount DECIMAL(10,2),
                                             currency VARCHAR(10),
                                             created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)