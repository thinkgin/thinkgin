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
-- Table structure for blog_article
-- ----------------------------
DROP TABLE IF EXISTS `blog_article`;
CREATE TABLE `blog_article` (
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
-- Records of blog_article
-- ----------------------------
INSERT INTO `blog_article` VALUES ('1', '1', 'test1', 'test-desc', 'test-content', '1648092878', 'test-created', '0', '', '0', '1');
INSERT INTO `blog_article` VALUES ('2', '1', 'test1', 'test-desc', 'test-content', '1648093530', 'test-created', '0', '', '0', '1');

-- ----------------------------
-- Table structure for blog_auth
-- ----------------------------
DROP TABLE IF EXISTS `blog_auth`;
CREATE TABLE `blog_auth` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(50) DEFAULT '' COMMENT '账号',
  `password` varchar(50) DEFAULT '' COMMENT '密码',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of blog_auth
-- ----------------------------

-- ----------------------------
-- Table structure for blog_tag
-- ----------------------------
DROP TABLE IF EXISTS `blog_tag`;
CREATE TABLE `blog_tag` (
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
-- Records of blog_tag
-- ----------------------------
INSERT INTO `blog_tag` VALUES ('1', 'news123', '1648048265', 'laowu', '0', '', '0', '1');
INSERT INTO `blog_tag` VALUES ('2', '2', '1648093612', 'test', '0', '', '0', '1');
INSERT INTO `blog_tag` VALUES ('3', 'goods33333', '1648093741', 'test', '1648449216', 'test12', '0', '1');
INSERT INTO `blog_tag` VALUES ('4', '1', '0', 'test', '0', '', '0', '1');
INSERT INTO `blog_tag` VALUES ('5', 'good', '0', 'test', '0', '', '0', '1');
INSERT INTO `blog_tag` VALUES ('6', 'goods', '0', 'test', '0', '', '0', '1');
INSERT INTO `blog_tag` VALUES ('7', 'goods2', '0', 'test', '0', '', '0', '1');
SET FOREIGN_KEY_CHECKS=1;
