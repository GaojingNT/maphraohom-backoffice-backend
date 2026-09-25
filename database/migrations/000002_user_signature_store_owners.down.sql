-- Restore tbl_stores.signature, copying each store's signature back from
-- one of its owners (lowest user id wins). Users' signatures and the
-- ownership table are kept — AutoMigrate would just recreate them anyway.
BEGIN;

ALTER TABLE tbl_stores ADD COLUMN IF NOT EXISTS signature varchar(255);

UPDATE tbl_stores s
SET signature = src.signature
FROM (
	SELECT DISTINCT ON (us.store_id) us.store_id, u.signature
	FROM tbl_user_stores us
	JOIN tbl_users u ON u.id = us.user_id
	WHERE COALESCE(u.signature, '') <> ''
	ORDER BY us.store_id, u.id
) src
WHERE s.id = src.store_id AND COALESCE(s.signature, '') = '';

COMMIT;
