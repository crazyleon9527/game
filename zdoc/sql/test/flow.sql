/*
 Navicat Premium Data Transfer

 Source Server         : cheetah_main
 Source Server Type    : MySQL
 Source Server Version : 50744
 Source Host           : cheetah.cpses0ockqsy.ap-south-1.rds.amazonaws.com:3306
 Source Schema         : cash_game_test

 Target Server Type    : MySQL
 Target Server Version : 50744
 File Encoding         : 65001

 Date: 31/03/2024 07:10:51
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for flow
-- ----------------------------
DROP TABLE IF EXISTS `flow`;
CREATE TABLE `flow`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` bigint(20) NULL DEFAULT NULL,
  `updated_at` bigint(20) NULL DEFAULT NULL,
  `uid` bigint(20) UNSIGNED NULL DEFAULT 0,
  `type` smallint(5) UNSIGNED NULL DEFAULT 0,
  `currency` varchar(12) CHARACTER SET utf8 COLLATE utf8_unicode_ci NULL DEFAULT NULL,
  `is_robot` tinyint(3) UNSIGNED NULL DEFAULT 0,
  `number` decimal(10, 2) NULL DEFAULT 0.00,
  `balance` decimal(11, 3) NULL DEFAULT 0.000,
  `remark` varchar(50) CHARACTER SET utf8 COLLATE utf8_unicode_ci NULL DEFAULT NULL,
  `pc` bigint(20) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_flow_uid`(`uid`) USING BTREE,
  INDEX `idx_flow_flow_type`(`type`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 162022 CHARACTER SET = utf8 COLLATE = utf8_unicode_ci ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;
