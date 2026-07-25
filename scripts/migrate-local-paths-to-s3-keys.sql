-- Rewrites podcast_items.download_path and podcast_items.local_image from
-- absolute local paths (/assets/<Podcast Name>/<file>) to the relative S3
-- object keys (<Podcast Name>/<file>) that the S3 backend expects, per
-- computeLocation() in service/fileService.go.
--
-- Run this ONLY after rclone has synced /assets/<Podcast Name>/<file> to
-- <bucket>/<Podcast Name>/<file> like-for-like, with no S3_KEY_PREFIX set
-- (the prefix, if any, is applied at read time by s3FullKey() - it must NOT
-- also be baked into these DB values, or every path will be doubled up).
--
-- Usage:
--   cp podgrab.db podgrab.db.bak                       # always keep a backup
--   sqlite3 podgrab.db.bak < scripts/preview-local-paths-to-s3-keys.sql   # optional dry run, see below
--   sqlite3 podgrab.db < scripts/migrate-local-paths-to-s3-keys.sql
--
-- The DATA prefix below must match your compose volume mount exactly
-- (this assumes the default '/assets', matching what's in this repo's
-- docker-compose.yml and what's actually stored in the sampled DB).

BEGIN;

UPDATE podcast_items
SET download_path = substr(download_path, length('/assets/') + 1)
WHERE download_path LIKE '/assets/%';

UPDATE podcast_items
SET local_image = substr(local_image, length('/assets/') + 1)
WHERE local_image LIKE '/assets/%';

-- Sanity checks - should both return 0 once this has run cleanly.
SELECT count(*) AS remaining_absolute_download_paths
FROM podcast_items WHERE download_path LIKE '/assets/%';

SELECT count(*) AS remaining_absolute_local_images
FROM podcast_items WHERE local_image LIKE '/assets/%';

-- Eyeball a few converted rows before trusting it.
SELECT id, download_path FROM podcast_items WHERE download_path != '' LIMIT 10;

COMMIT;
