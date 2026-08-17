CREATE TABLE IF NOT EXISTS tickets(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  type VARCHAR(255) NOT NULL,
  price NUMERIC(10, 2) NOT NULL,
  quantity_total INT NOT NULL,
  quantity_available INT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now() NOT NULL
);