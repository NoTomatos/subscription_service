CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY,
    
    service_name VARCHAR(255) NOT NULL,
    
    price INTEGER NOT NULL CHECK (price >= 0),
    
    user_id UUID NOT NULL,
    
    start_date DATE NOT NULL,
    
    end_date DATE,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_service_name ON subscriptions(service_name);
CREATE INDEX idx_subscriptions_dates ON subscriptions(start_date, end_date);
CREATE INDEX idx_subscriptions_user_dates ON subscriptions(user_id, start_date, end_date);
