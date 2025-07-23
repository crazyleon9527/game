/*
 Navicat Premium Data Transfer

 Source Server         : localhost
 Source Server Type    : MySQL
 Source Server Version : 80012
 Source Host           : localhost:3306
 Source Schema         : cash_game

 Target Server Type    : MySQL
 Target Server Version : 80012
 File Encoding         : 65001

 Date: 30/04/2024 17:51:29
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for recharge_setting
-- ----------------------------
DROP TABLE IF EXISTS `recharge_setting`;
CREATE TABLE `recharge_setting`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` bigint(20) NULL DEFAULT NULL,
  `updated_at` bigint(20) NULL DEFAULT NULL,
  `deleted_at` datetime(0) NULL DEFAULT NULL,
  `name` varchar(20) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `recharge_state` tinyint(3) UNSIGNED NULL DEFAULT 0,
  `withdraw_state` tinyint(3) UNSIGNED NULL DEFAULT 0,
  `sort` tinyint(3) UNSIGNED NULL DEFAULT NULL,
  `status` tinyint(3) UNSIGNED NULL DEFAULT 0,
  `available_amount` decimal(10, 2) NULL DEFAULT 0.00,
  `frozen_amount` decimal(10, 2) NULL DEFAULT 0.00,
  `balance_amount` decimal(10, 2) NULL DEFAULT 0.00,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_recharge_setting_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 6 CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of recharge_setting
-- ----------------------------
INSERT INTO `recharge_setting` VALUES (1, 1708463353, 1708463353, NULL, 'poly', 0, 0, NULL, 0, 0.00, 0.00, 0.00);
INSERT INTO `recharge_setting` VALUES (2, 1708463353, 1714470443, NULL, 'kb', 1, 0, 1, 1, 862216.43, 21636.05, 883852.48);
INSERT INTO `recharge_setting` VALUES (3, 1708463353, 1714470443, NULL, 'tk', 1, 0, 2, 1, 35030.33, 0.00, 44747.48);
INSERT INTO `recharge_setting` VALUES (4, 1708463353, 1711964798, NULL, 'at', 0, 1, 3, 1, 349319.19, 14469.90, 363789.09);
INSERT INTO `recharge_setting` VALUES (5, 1708463353, 1714470446, NULL, 'go', 1, 0, 0, 1, 35482.73, 0.00, 35682.73);
INSERT INTO `recharge_setting` VALUES (6, 1708463353, 1714414598, NULL, 'cow', 1, 1, 4, 1, 0.00, 0.00, 0.00);

SET FOREIGN_KEY_CHECKS = 1;
