ALTER TABLE products ADD COLUMN IF NOT EXISTS image_path_2 text DEFAULT '' NOT NULL;
ALTER TABLE products ADD COLUMN IF NOT EXISTS image_path_3 text DEFAULT '' NOT NULL;

UPDATE products
SET
    image_path_2 = '/static/assets/images/gallery/car-washer/car-washer2.jpg',
    image_path_3 = '/static/assets/images/gallery/car-washer/car-washer3.png'
WHERE sku = 'CAR-WASHER-001';
