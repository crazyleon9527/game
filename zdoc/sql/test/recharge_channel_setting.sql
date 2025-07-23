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

 Date: 30/04/2024 17:51:16
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for recharge_channel_setting
-- ----------------------------
DROP TABLE IF EXISTS `recharge_channel_setting`;
CREATE TABLE `recharge_channel_setting`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `pcid` bigint(20) UNSIGNED NULL DEFAULT 0,
  `created_at` bigint(20) NULL DEFAULT NULL,
  `updated_at` bigint(20) NULL DEFAULT NULL,
  `name` varchar(20) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `app_id` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `pay_key` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `pay_secret` varchar(1000) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `recharge_api_url` varchar(256) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `pay_callback_url` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `pay_return_url` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `sett_callback_url` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `withdraw_key` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `sett_return_url` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `withdraw_api_url` varchar(256) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `remark` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  `status` tinyint(3) UNSIGNED NULL DEFAULT 0,
  `deleted_at` datetime(0) NULL DEFAULT NULL,
  `balance_api_url` varchar(256) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_recharge_channel_setting_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 6 CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of recharge_channel_setting
-- ----------------------------
INSERT INTO `recharge_channel_setting` VALUES (1, 1, 1708463353, 1708463353, 'poly', '861100000013114', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 1, NULL, NULL);
INSERT INTO `recharge_channel_setting` VALUES (2, 2, 1708463353, 1708463353, 'kb', '6058632', 'KEUGFymu6esRaEnFQlvjMKXd4dRlyP5D', NULL, 'https://api.kbpay.io/debug/payin/submit', 'https://api-dev.cheetahs.asia/api/recharge/callback/kb', 'https://fc-dev.cheetahs.asia/api/recharge/index.html', 'https://api.cheetahs.asia/api/withdraw/callback/kb', 'db383PIA9uFsOCJl9X8xccaxw2XgOIY7', NULL, 'https://api.kbpay.io/debug/payout/submit', NULL, 1, NULL, 'https://api.kbpay.io/balance');
INSERT INTO `recharge_channel_setting` VALUES (3, 3, 1708463353, 1708463353, 'tk', '202366100', 'c385fe7029344aef826d8112625b2625', NULL, 'https://seabird.world/api/order/pay/create', 'https://api-dev.cheetahs.asia/api/recharge/callback/tk', 'https://fc-dev.cheetahs.asia/api/recharge/index.html', 'https://api-dev.cheetahs.asia/api/withdraw/callback/tk', 'c385fe7029344aef826d8112625b2625', NULL, 'https://seabird.world/api/order/withdraw/create', NULL, 1, NULL, 'https://seabird.world/api/bal');
INSERT INTO `recharge_channel_setting` VALUES (4, 4, 1708463353, 1708463353, 'at', '563186', 'D6BC07JDRGCKTPJI6IZHUKF0', 'DN2DOMLMVCM69NNHEYDPJVW19ZEU07QL', 'https://api.atpayment.co/trade/v1/unifiedorder/legal', 'https://api-dev.cheetahs.asia/api/recharge/callback/at', 'https://fc-dev.cheetahs.asia/api/recharge/index.html', 'https://api-dev.cheetahs.asia/api/withdraw/callback/at', 'D6BC07JDRGCKTPJI6IZHUKF0', NULL, 'https://api.atpayment.co/trade/v1/agentpay/legal', NULL, 1, NULL, 'https://api.atpayment.co/mer/v1/balanceQuery');
INSERT INTO `recharge_channel_setting` VALUES (5, 5, 1708463353, 1708463353, 'go', '2023100001', '2c045601744c4b149d5b0f51dec1dbd3', NULL, 'https://gooopay.online/api/recharge/create', 'https://api-dev.cheetahs.asia/api/recharge/callback/go', 'dev.cheetahs.asia/api/recharge/index.html', 'https://api-dev.cheetahs.asia/api/withdraw/callback/go', '2c045601744c4b149d5b0f51dec1dbd3', NULL, 'https://gooopay.online/api/deposit/create', NULL, 1, NULL, 'https://gooopay.online/api/balance');
INSERT INTO `recharge_channel_setting` VALUES (6, 6, 1708463353, 1708463353, 'cow', '1714292212272', '43bd77a845b238ae5a300f075ef89027', NULL, 'https://pay365.cowpay.co/pay', 'https://api-dev.cheetahs.asia/api/recharge/callback/cow', 'dev.cheetahs.asia/api/recharge/index.html', 'dev.cheetahs.asia/api/withdraw/callback/cow', NULL, NULL, 'https://pay365.cowpay.co/v2/withdraw', NULL, 1, NULL, 'https://pay365.cowpay.co/v2/queryBalance');

SET FOREIGN_KEY_CHECKS = 1;
