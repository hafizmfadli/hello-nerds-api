CREATE TABLE `order_items` (
  `id` int NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `order_id` int NOT NULL,
  `updated_edited_id` int unsigned NOT NULL,
  `quantity` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `fk_order_items_orders` (`order_id`),
  KEY `fk_order_items_updated_edited` (`updated_edited_id`),
  CONSTRAINT `fk_order_items_updated_edited` FOREIGN KEY (`updated_edited_id`) REFERENCES `updated_edited` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_order_items_orders` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB;