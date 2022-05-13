CREATE TABLE IF NOT EXISTS tokens (
  hash VARBINARY(255) PRIMARY KEY,
  expiry TIMESTAMP NOT NULL,
  scope VARCHAR(255) NOT NULL,
  user_id INT NOT NULL,
  CONSTRAINT fk_tokens_users
  FOREIGN KEY (user_id)   
  REFERENCES users(id) 
    ON DELETE CASCADE
);