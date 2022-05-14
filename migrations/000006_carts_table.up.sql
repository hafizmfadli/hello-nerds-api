CREATE TABLE IF NOT EXISTS `carts` (
  `id` int NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `user_id` int NOT NULL,
  `updated_edited_id` int unsigned NOT NULL,
  `quantity` int NOT NULL DEFAULT '0',
  `total_price` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `fk_carts_users` (`user_id`),
  KEY `fk_carts_updated_edited` (`updated_edited_id`),
  CONSTRAINT `fk_carts_updated_edited` FOREIGN KEY (`updated_edited_id`) REFERENCES `updated_edited` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_carts_users` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB;