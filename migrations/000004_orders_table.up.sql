CREATE TABLE IF NOT EXISTS `orders` (
  `id` int NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `user_id` int,
  `shipping_address_id` int NOT NULL,
  `is_paid` tinyint(1) NOT NULL DEFAULT '0',
  `payment_deadline` timestamp NOT NULL,
  `total_price` BIGINT DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `fk_orders_users` (`user_id`),
  KEY `fk_orders_shipping_address` (`shipping_address_id`),
  CONSTRAINT `fk_orders_shipping_address` FOREIGN KEY (`shipping_address_id`) REFERENCES `shipping_address` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_orders_users` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
)