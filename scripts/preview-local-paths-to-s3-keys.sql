-- Read-only preview for migrate-local-paths-to-s3-keys.sql - makes no
-- changes. Run first to confirm the before/after mapping looks right:
--   sqlite3 podgrab.db < scripts/preview-local-paths-to-s3-keys.sql

SELECT count(*) AS download_paths_to_migrate
FROM podcast_items WHERE download_path LIKE '/assets/%';

SELECT count(*) AS local_images_to_migrate
FROM podcast_items WHERE local_image LIKE '/assets/%';

SELECT count(*) AS download_paths_NOT_matching_assets_prefix
FROM podcast_items WHERE download_path != '' AND download_path NOT LIKE '/assets/%';

SELECT count(*) AS local_images_NOT_matching_assets_prefix
FROM podcast_items WHERE local_image != '' AND local_image NOT LIKE '/assets/%';

-- Sample before -> after for a few rows.
SELECT
  download_path AS before,
  substr(download_path, length('/assets/') + 1) AS after
FROM podcast_items
WHERE download_path LIKE '/assets/%'
LIMIT 10;
