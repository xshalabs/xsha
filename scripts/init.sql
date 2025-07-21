-- MySQL database initialization script
-- Create database (if not exists)
CREATE DATABASE IF NOT EXISTS sleep0 CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Use database
USE sleep0;

-- Set timezone
SET time_zone = '+08:00';

-- Create user table (example)
-- Note: Actual table structure will be automatically created and migrated by GORM
-- This is just to ensure database connection is working

-- Insert some initial data (optional)
-- INSERT INTO users (username, email, created_at, updated_at) VALUES
-- ('admin', 'admin@example.com', NOW(), NOW());

-- Set permissions
GRANT ALL PRIVILEGES ON sleep0.* TO 'sleep0'@'%';
FLUSH PRIVILEGES;

-- Show created databases
SHOW DATABASES;
SHOW TABLES; 