CREATE TABLE IF NOT EXISTS shipping_address (
  id INT NOT NULL AUTO_INCREMENT,
  email VARCHAR(255) NOT NULL,
  first_name VARCHAR(255) NOT NULL,
  last_name VARCHAR(255) NOT NULL,
  addresses VARCHAR(255) NOT NULL,
  postal_code VARCHAR(10) NOT NULL,
  province_id INT NOT NULL,
  city_id INT NOT NULL,
  district_id INT NOT NULL,
  subdistrict_id INT NOT NULL,
  phone VARCHAR(20) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  user_id INT,
  CONSTRAINT fk_shipping_address_users
  FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE,
  PRIMARY KEY (id)
);