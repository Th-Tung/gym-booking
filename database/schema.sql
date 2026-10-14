-- Gym Equipment Booking System Database Schema (Go Backend)
-- Charset: utf8mb4

CREATE DATABASE IF NOT EXISTS `gym_booking` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `gym_booking`;

-- Table: users
CREATE TABLE IF NOT EXISTS `users` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL,
    `email` VARCHAR(150) NOT NULL UNIQUE,
    `password_hash` VARCHAR(255) NOT NULL,
    `role` ENUM('admin', 'member') NOT NULL DEFAULT 'member',
    `phone` VARCHAR(20) NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: categories
CREATE TABLE IF NOT EXISTS `categories` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL,
    `description` TEXT NULL,
    `icon` VARCHAR(50) DEFAULT 'bi-collection',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: equipments
CREATE TABLE IF NOT EXISTS `equipments` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `category_id` INT NOT NULL,
    `name` VARCHAR(150) NOT NULL,
    `description` TEXT NULL,
    `image_path` VARCHAR(255) NULL,
    `status` ENUM('available', 'maintenance') NOT NULL DEFAULT 'available',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT `fk_equipment_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: bookings
CREATE TABLE IF NOT EXISTS `bookings` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `user_id` INT NOT NULL,
    `equipment_id` INT NOT NULL,
    `start_time` DATETIME NOT NULL,
    `end_time` DATETIME NOT NULL,
    `status` ENUM('pending', 'confirmed', 'cancelled') NOT NULL DEFAULT 'confirmed',
    `notes` TEXT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT `fk_booking_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_booking_equipment` FOREIGN KEY (`equipment_id`) REFERENCES `equipments` (`id`) ON DELETE CASCADE,
    INDEX `idx_equipment_time` (`equipment_id`, `start_time`, `end_time`),
    INDEX `idx_user_bookings` (`user_id`, `start_time`)
-- Table: maintenance_reports
CREATE TABLE IF NOT EXISTS `maintenance_reports` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `equipment_id` INT NOT NULL,
    `user_id` INT NOT NULL,
    `issue_description` TEXT NOT NULL,
    `status` ENUM('pending', 'confirmed_maintenance', 'resolved', 'rejected') NOT NULL DEFAULT 'pending',
    `admin_notes` TEXT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `resolved_at` TIMESTAMP NULL,
    CONSTRAINT `fk_report_equipment` FOREIGN KEY (`equipment_id`) REFERENCES `equipments` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_report_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Initial Seed Data
-- Default Admin: admin@gym.com / Admin@1234
-- Default Member: member@gym.com / Member@1234
INSERT INTO `users` (`id`, `name`, `email`, `password_hash`, `role`, `phone`) VALUES
(1, 'Gym Administrator', 'admin@gym.com', '$2a$10$Q7eY8PqXqE1K7y3Q9H0V4.mRkI/gRkRz0V3E8U4M8re5181v929L7', 'admin', '0812345678'),
(2, 'John Workout', 'member@gym.com', '$2a$10$wN2L0xYdC5.rW4U4E8m8re5181v929L7Z0M3Xy32P8oV24T7R36e2', 'member', '0898765432')
ON DUPLICATE KEY UPDATE `email`=`email`;

-- Seed Categories
INSERT INTO `categories` (`id`, `name`, `description`, `icon`) VALUES
(1, 'Cardio Zone', 'อุปกรณ์เพิ่มอัตราการเต้นหัวใจ เผาผลาญไขมัน เช่น ลู่วิ่ง จักรยาน และเครื่องเดินวงรี', 'bi-heart-pulse'),
(2, 'Free Weights', 'ดัมเบล บาร์เบล และแผ่นน้ำหนัก เพื่อการฝึกกล้ามเนื้อเฉพาะส่วนอย่างอิสระ', 'bi-speedometer2'),
(3, 'Resistance Machines', 'เครื่องออกกำลังกายแบบมีรอกและหมุดปรับน้ำหนัก ปลอดภัย เหมาะกับทุกระดับ', 'bi-cpu'),
(4, 'Functional & Cable', 'เครื่องเคเบิลครอสโอเวอร์และอุปกรณ์เทรนนิ่งอเนกประสงค์', 'bi-activity')
ON DUPLICATE KEY UPDATE `name`=`name`;

-- Seed Equipments
INSERT INTO `equipments` (`id`, `category_id`, `name`, `description`, `image_path`, `status`) VALUES
(1, 1, 'Commercial Treadmill T900', 'ลู่วิ่งไฟฟ้าระดับมืออาชีพ พร้อมหน้าจอทัชสกรีนและโหมดวิ่งขึ้นเขาปรับระดับได้', 'assets/images/treadmill.svg', 'available'),
(2, 1, 'Stationary Spin Bike Pro', 'จักรยานปั่นสปินนิ่งพร้อมระบบวัดพลังวัตต์และแรงต้านแม่เหล็ก ปั่นลื่นเงียบสนิท', 'assets/images/spin_bike.svg', 'available'),
(3, 1, 'Elliptical Trainer E80', 'เครื่องเดินวงรีลดแรงกระแทกข้อต่อ เหมาะสำหรับการคาร์ดิโอถนอมเข่า', 'assets/images/elliptical.svg', 'available'),
(4, 2, 'Olympic Bench Press Rack', 'ม้านอนยกบาร์เบลมาตรฐานโอลิมปิก พร้อมเซฟตี้บาร์เพื่อความปลอดภัยสูงสุด', 'assets/images/bench_press.svg', 'available'),
(5, 2, 'Power Cage / Squat Rack', 'กรงสควอทอเนกประสงค์พร้อมแท่นดึงข้อและบาร์ปรับระดับได้สำหรับหนักหน่วง', 'assets/images/squat_rack.svg', 'available'),
(6, 3, 'Leg Press Machine 45°', 'เครื่องบริหารต้นขาและสะโพกแบบ 45 องศา รองรับน้ำหนักสูงสุด 450 กก.', 'assets/images/leg_press.svg', 'available'),
(7, 3, 'Chest Press Machine', 'เครื่องออกกำลังกายกล้ามเนื้อหน้าอกแบบแยกแขนอิสระ โฟกัสกล้ามเนื้อได้ตรงจุด', 'assets/images/chest_press.svg', 'maintenance'),
(8, 4, 'Dual Cable Crossover', 'เครื่องเคเบิลคู่ปรับระดับความสูงได้รอบทิศทาง ฝึกได้ครบทุกส่วนของร่างกาย', 'assets/images/cable_crossover.svg', 'available')
ON DUPLICATE KEY UPDATE `name`=`name`;
