-- Clean all business tables and reset sequences
TRUNCATE TABLE
    tbl_bill_items,
    tbl_bills,
    tbl_store_product_prices,
    tbl_customer_addresses,
    tbl_customers,
    tbl_products,
    tbl_stores
RESTART IDENTITY CASCADE;

-- Products (shared across stores)
INSERT INTO tbl_products (name, created_at, updated_at) VALUES
    ('เนื้อหั่นเส้น',  NOW(), NOW()),
    ('เนื้อหั่นชิ้น',  NOW(), NOW()),
    ('เนื้อเบ้าพับ',   NOW(), NOW()),
    ('เนื้อขูดฝอย',   NOW(), NOW()),
    ('เนื้อริ้ว',      NOW(), NOW()),
    ('น้ำมะพร้าว',    NOW(), NOW());

-- Stores
INSERT INTO tbl_stores (name, created_at, updated_at) VALUES
    ('มะพร้าวหอมแปรรูป อัมพวา',         NOW(), NOW()),
    ('Maphraohom Ampawa มะพร้าวหอมอัมพวา', NOW(), NOW());

-- Store-product prices
-- Store 1: มะพร้าวหอมแปรรูป อัมพวา (id=1)  → 80/80/80/80/80/50
-- Store 2: Maphraohom Ampawa (id=2)          → 100/100/100/100/100/70
INSERT INTO tbl_store_product_prices (store_id, product_id, price, effective_from, created_at, updated_at) VALUES
    (1, 1, 80,  NOW(), NOW(), NOW()),
    (1, 2, 80,  NOW(), NOW(), NOW()),
    (1, 3, 80,  NOW(), NOW(), NOW()),
    (1, 4, 80,  NOW(), NOW(), NOW()),
    (1, 5, 80,  NOW(), NOW(), NOW()),
    (1, 6, 50,  NOW(), NOW(), NOW()),
    (2, 1, 100, NOW(), NOW(), NOW()),
    (2, 2, 100, NOW(), NOW(), NOW()),
    (2, 3, 100, NOW(), NOW(), NOW()),
    (2, 4, 100, NOW(), NOW(), NOW()),
    (2, 5, 100, NOW(), NOW(), NOW()),
    (2, 6, 70,  NOW(), NOW(), NOW());
