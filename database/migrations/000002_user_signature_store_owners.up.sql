-- Move the printed signature from stores to users, and add store ownership
-- (many-to-many: a store can have several owners, an owner several stores).
--
-- Run order on an existing deployment:
--   1. Deploy + start the app once — GORM AutoMigrate creates
--      tbl_user_stores and tbl_users.signature (this file also does, so
--      order doesn't break anything, it just matters for step 2).
--   2. Assign owners (INSERT INTO tbl_user_stores ...) if you want each
--      store's existing signature carried over to its owners.
--   3. Run this migration. Every owner whose own signature is still empty
--      gets their store's signature copied over (lowest store id wins when
--      they own several), then tbl_stores.signature is dropped.
--
-- A store signature with no owner at that point is NOT carried anywhere —
-- the object stays in storage but nothing references it any more.
--
-- Idempotent: re-running after a successful run is a no-op.
BEGIN;

ALTER TABLE tbl_users ADD COLUMN IF NOT EXISTS signature varchar(255);

CREATE TABLE IF NOT EXISTS tbl_user_stores (
	user_id bigint NOT NULL REFERENCES tbl_users (id) ON UPDATE CASCADE ON DELETE CASCADE,
	store_id bigint NOT NULL REFERENCES tbl_stores (id) ON UPDATE CASCADE ON DELETE CASCADE,
	PRIMARY KEY (user_id, store_id)
);

CREATE INDEX IF NOT EXISTS idx_tbl_user_stores_store_id ON tbl_user_stores (store_id);

DO $$ BEGIN
	IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'tbl_stores' AND column_name = 'signature') THEN
		UPDATE tbl_users u
		SET signature = src.signature
		FROM (
			SELECT DISTINCT ON (us.user_id) us.user_id, s.signature
			FROM tbl_user_stores us
			JOIN tbl_stores s ON s.id = us.store_id
			WHERE COALESCE(s.signature, '') <> ''
			ORDER BY us.user_id, s.id
		) src
		WHERE u.id = src.user_id AND COALESCE(u.signature, '') = '';

		ALTER TABLE tbl_stores DROP COLUMN signature;
	END IF;
END $$;

COMMIT;
