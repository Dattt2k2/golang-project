ALTER TABLE orders ALTER COLUMN order_id TYPE UUID USING order_id::uuid;
ALTER TABLE orders ALTER COLUMN order_id SET DEFAULT gen_random_uuid();