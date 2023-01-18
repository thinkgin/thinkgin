/*
Navicat MySQL Data Transfer

Source Server         : Go2022
Source Server Version : 50648
Source Host           : 110.42.217.11:3306
Source Database       : go2022

Target Server Type    : MYSQL
Target Server Version : 50648
File Encoding         : 65001

Date: 2022-03-29 10:06:20
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for tg_article
-- ----------------------------
DROP TABLE IF EXISTS `tg_article`;
CREATE TABLE `tg_article` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `tag_id` int(10) unsigned DEFAULT '0' COMMENT '标签ID',
  `title` varchar(100) DEFAULT '' COMMENT '文章标题',
  `desc` varchar(255) DEFAULT '' COMMENT '简述',
  `content` text,
  `created_on` int(11) DEFAULT NULL,
  `created_by` varchar(100) DEFAULT '' COMMENT '创建人',
  `modified_on` int(10) unsigned DEFAULT '0' COMMENT '修改时间',
  `modified_by` varchar(255) DEFAULT '' COMMENT '修改人',
  `deleted_on` int(10) unsigned DEFAULT '0',
  `state` tinyint(3) unsigned DEFAULT '1' COMMENT '状态 0为禁用1为启用',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8 COMMENT='文章管理';

-- ----------------------------
-- Records of tg_article
-- ----------------------------
INSERT INTO `tg_article` VALUES ('1', '1', 'test1', 'test-desc', 'test-content', '1648092878', 'test-created', '0', '', '0', '1');
INSERT INTO `tg_article` VALUES ('2', '1', 'test1', 'test-desc', 'test-content', '1648093530', 'test-created', '0', '', '0', '1');

-- ----------------------------
-- Table structure for tg_auth
-- ----------------------------
DROP TABLE IF EXISTS `tg_auth`;
CREATE TABLE `tg_auth` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(50) DEFAULT '' COMMENT '账号',
  `password` varchar(50) DEFAULT '' COMMENT '密码',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of tg_auth
-- ----------------------------

-- ----------------------------
-- Table structure for tg_tag
-- ----------------------------
DROP TABLE IF EXISTS `tg_tag`;
CREATE TABLE `tg_tag` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) DEFAULT '' COMMENT '标签名称',
  `created_on` int(10) unsigned DEFAULT '0' COMMENT '创建时间',
  `created_by` varchar(100) DEFAULT '' COMMENT '创建人',
  `modified_on` int(10) unsigned DEFAULT '0' COMMENT '修改时间',
  `modified_by` varchar(100) DEFAULT '' COMMENT '修改人',
  `deleted_on` int(10) unsigned DEFAULT '0',
  `state` tinyint(3) unsigned DEFAULT '1' COMMENT '状态 0为禁用、1为启用',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8 COMMENT='文章标签管理';

-- ----------------------------
-- Records of tg_tag
-- ----------------------------
INSERT INTO `tg_tag` VALUES ('1', 'laowutest', '1648048265', 'laowu', '0', '', '0', '1');
INSERT INTO `tg_tag` VALUES ('2', 'thinkgin', '1648093612', 'laowu', '0', '', '0', '1');
SET FOREIGN_KEY_CHECKS=1;
