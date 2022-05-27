CREATE DEFINER=`root`@`%` PROCEDURE `periplus_dev`.`checkout_v1`(array_of_book_id_param VARCHAR(1000), array_of_book_quantity_param VARCHAR(1000), shipping_email_param VARCHAR(255), shipping_first_name_param VARCHAR(255), shipping_last_name_param VARCHAR(255), shipping_addresses_param VARCHAR(255),
shipping_postal_code_param VARCHAR(10), shipping_province_id_param INT, shipping_city_id_param INT, shipping_district_id_param INT, shipping_subdistrict_id_param INT, phone_param VARCHAR(20), 
user_id_param INT, checkout_variety_param INT, address_variety_param INT, existing_shipping_address_param INT)
BEGIN
-- 	constant
	DECLARE GUEST_CHECKOUT INT DEFAULT 0;
	DECLARE MEMBER_CHECKOUT INT DEFAULT 1;
	DECLARE TO_NEW_ADDRESS INT DEFAULT 0;
	DECLARE TO_EXISTING_ADDRESS INT DEFAULT 1;
	
	DECLARE user_id INT DEFAULT NULL;
	DECLARE shipping_address_id INT;
	DECLARE order_id INT;
	DECLARE array_of_book_id VARCHAR(1000);
	DECLARE array_of_book_quantity VARCHAR(1000);
	DECLARE is_enough TINYINT;
	DECLARE book_price INT;
	DECLARE total_order_price BIGINT DEFAULT 0;

	DECLARE start_pos SMALLINT;
	DECLARE comma_pos SMALLINT;
	DECLARE current_id INT;
	DECLARE current_quantity INT;
	DECLARE end_loop TINYINT;
	

	SET array_of_book_id = array_of_book_id_param;
	SET array_of_book_quantity = array_of_book_quantity_param;
	SET start_pos = 1;
	SET comma_pos = LOCATE(',', array_of_book_id);

	START TRANSACTION;
	
	IF checkout_variety_param = MEMBER_CHECKOUT THEN
		SET user_id = user_id_param;
	END IF;

	IF address_variety_param = TO_NEW_ADDRESS THEN
		-- ship to new address		
		-- create new shipping address
			INSERT INTO shipping_address(email, first_name, last_name, addresses, postal_code, 
			province_id, city_id, district_id, subdistrict_id, phone, user_id) VALUES (shipping_email_param, shipping_first_name_param,
			shipping_last_name_param, shipping_addresses_param, shipping_postal_code_param, shipping_province_id_param, shipping_city_id_param,
			shipping_district_id_param, shipping_subdistrict_id_param, phone_param, user_id);
			
			SELECT id INTO shipping_address_id FROM shipping_address ORDER BY id DESC LIMIT 1;
		
	ELSEIF address_variety_param = TO_EXISTING_ADDRESS THEN
		-- use existing shipping address
		SET shipping_address_id = existing_shipping_address_param;
	END IF;
	
	-- create a new order
		INSERT INTO orders(user_id, shipping_address_id, is_paid, payment_deadline, total_price) VALUES(user_id, shipping_address_id, FALSE, NOW(), 0);	
		SELECT id INTO order_id FROM orders ORDER BY id DESC LIMIT 1;

	REPEAT
		IF comma_pos > 0 THEN
			SET current_id = CAST(SUBSTRING(array_of_book_id, start_pos, comma_pos - start_pos) AS UNSIGNED);
			SET current_quantity = CAST(SUBSTRING(array_of_book_quantity, start_pos, comma_pos - start_pos) AS UNSIGNED);
			SET end_loop = FALSE;
		ELSE
			SET current_id = CAST(SUBSTRING(array_of_book_id, start_pos) AS UNSIGNED);
			SET current_quantity = CAST(SUBSTRING(array_of_book_quantity, start_pos) AS UNSIGNED);
			SET end_loop = TRUE;
		END IF;
		
		-- check is stock is enough for buy by user ?
		SELECT quantity >= current_quantity, price INTO is_enough, book_price FROM updated_edited ue WHERE ue.id = current_id FOR UPDATE;
		
	
		-- if stock is not enough then rollback and throw an exception
		IF is_enough = FALSE THEN
			ROLLBACK;
		SIGNAL SQLSTATE '45000'
			SET MESSAGE_TEXT = 'Not enough stock';
		END IF;
		
		-- substract book stock with quantity book tha user will be buy
		UPDATE updated_edited SET quantity = quantity - current_quantity WHERE id = current_id;
	
		-- insert order items
		INSERT INTO order_items(order_id, updated_edited_id, quantity, total_price) VALUES(order_id, current_id, current_quantity, book_price * current_quantity);
		
		-- accumulate total order price		
		SET total_order_price = total_order_price + book_price * current_quantity;
		
		IF end_loop = 0 THEN
			SET array_of_book_id = SUBSTRING(array_of_book_id, comma_pos + 1);
			SET array_of_book_quantity = SUBSTRING(array_of_book_quantity, comma_pos + 1);
			SET comma_pos = LOCATE(',', array_of_book_id);
		END IF;
	
	UNTIL end_loop = TRUE
	END REPEAT;
	
	-- update order total price
	UPDATE orders SET total_price = total_order_price WHERE id = order_id; 

	COMMIT;

END