-- Drop legacy desired/observed image columns if present
ALTER TABLE tenants
  DROP COLUMN IF EXISTS desired_image,
  DROP COLUMN IF EXISTS observed_image;
