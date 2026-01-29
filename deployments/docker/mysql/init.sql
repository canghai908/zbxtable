-- ZbxTable 数据库初始化脚本
-- 此脚本会在 MySQL 容器首次启动时自动执行

-- 设置字符集
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS zbxtable DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE zbxtable;

-- 注意：表结构会在应用首次启动时自动创建
-- 这里只需要确保数据库存在即可

SET FOREIGN_KEY_CHECKS = 1;

