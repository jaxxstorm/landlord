-- Reintroduce legacy desired/observed image columns for rollback
ALTER TABLE tenants
  ADD COLUMN IF NOT EXISTS desired_image VARCHAR(500) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS observed_image VARCHAR(500) NOT NULL DEFAULT '';
