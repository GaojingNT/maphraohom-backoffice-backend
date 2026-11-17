-- Sales platforms (Facebook, TikTok, Lazada, Shopee, Thaimart) a receipt
-- bill can be tagged with: a new tbl_platform_types table and a nullable
-- tbl_bills.platform_type_id pointing at it.
--
-- The app does all of this on start (GORM AutoMigrate + 
-- pkg/database/platform_type_seed.go, which also uploads the logos to
-- storage), so on a normal deployment this file is a no-op; it is kept for
-- applying the change by hand. Logos are NOT uploaded by this file.
--
-- Additive only: existing bills get platform_type_id NULL, no bill row is
-- changed. Idempotent.
BEGIN;

CREATE TABLE IF NOT EXISTS tbl_platform_types (
	id         bigserial PRIMARY KEY,
	created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
	code       varchar(30)  NOT NULL,
	name       varchar(100) NOT NULL,
	logo       varchar(255),
	sort_order bigint  NOT NULL DEFAULT 0,
	is_active  boolean NOT NULL DEFAULT true
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tbl_platform_types_code ON tbl_platform_types (code);

INSERT INTO tbl_platform_types (code, name, logo, sort_order) VALUES
	('facebook', 'Facebook', 'platform-types/facebook.png', 1),
	('tiktok',   'TikTok',   'platform-types/tiktok.jpg',   2),
	('lazada',   'Lazada',   'platform-types/lazada.jpg',   3),
	('shopee',   'Shopee',   'platform-types/shopee.png',   4),
	('thaimart', 'Thaimart', 'platform-types/thaimart.png', 5)
ON CONFLICT (code) DO NOTHING;

ALTER TABLE tbl_bills ADD COLUMN IF NOT EXISTS platform_type_id bigint;
CREATE INDEX IF NOT EXISTS idx_tbl_bills_platform_type_id ON tbl_bills (platform_type_id);

DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tbl_bills_platform_type') THEN
		ALTER TABLE tbl_bills ADD CONSTRAINT fk_tbl_bills_platform_type
			FOREIGN KEY (platform_type_id) REFERENCES tbl_platform_types (id)
			ON UPDATE CASCADE ON DELETE RESTRICT;
	END IF;
END $$;

COMMIT;
