CREATE TABLE IF NOT EXISTS permissions (
  id INT NOT NULL AUTO_INCREMENT,
  code VARCHAR(255) NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS users_permissions (
  user_id INT NOT NULL,
  permission_id INT NOT NULL,
  CONSTRAINT fk_users_permissions_users
  FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE,
  CONSTRAINT fk_users_permissions_permissions
  FOREIGN KEY (permission_id)
    REFERENCES permissions(id)
    ON DELETE CASCADE,
  PRIMARY KEY (user_id, permission_id)
);

-- Add the two permission to the table.
INSERT INTO permissions (code) VALUES ('books:read'), ('books:write');