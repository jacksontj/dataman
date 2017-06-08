/*
Navicat PGSQL Data Transfer

Source Server         : local postgres
Source Server Version : 90506
Source Host           : localhost:5432
Source Database       : dataman_router
Source Schema         : public

Target Server Type    : PGSQL
Target Server Version : 90506
File Encoding         : 65001

Date: 2017-06-07 18:44:30
*/


-- ----------------------------
-- Sequence structure for collection__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection__id_seq";
CREATE SEQUENCE "public"."collection__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 4092
 CACHE 1;
SELECT setval('"public"."collection__id_seq"', 4092, true);

-- ----------------------------
-- Sequence structure for collection_field__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_field__id_seq";
CREATE SEQUENCE "public"."collection_field__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 17011
 CACHE 1;
SELECT setval('"public"."collection_field__id_seq"', 17011, true);

-- ----------------------------
-- Sequence structure for collection_field_relation__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_field_relation__id_seq";
CREATE SEQUENCE "public"."collection_field_relation__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1219
 CACHE 1;
SELECT setval('"public"."collection_field_relation__id_seq"', 1219, true);

-- ----------------------------
-- Sequence structure for collection_index__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_index__id_seq";
CREATE SEQUENCE "public"."collection_index__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 5320
 CACHE 1;
SELECT setval('"public"."collection_index__id_seq"', 5320, true);

-- ----------------------------
-- Sequence structure for collection_index_item__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_index_item__id_seq";
CREATE SEQUENCE "public"."collection_index_item__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 60353
 CACHE 1;
SELECT setval('"public"."collection_index_item__id_seq"', 60353, true);

-- ----------------------------
-- Sequence structure for collection_partition_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_partition_id_seq";
CREATE SEQUENCE "public"."collection_partition_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 4088
 CACHE 1;
SELECT setval('"public"."collection_partition_id_seq"', 4088, true);

-- ----------------------------
-- Sequence structure for collection_vshard__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_vshard__id_seq";
CREATE SEQUENCE "public"."collection_vshard__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for collection_vshard_instance__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_vshard_instance__id_seq";
CREATE SEQUENCE "public"."collection_vshard_instance__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for collection_vshard_instance_datastore_shard__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_vshard_instance_datastore_shard__id_seq";
CREATE SEQUENCE "public"."collection_vshard_instance_datastore_shard__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for constraint__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."constraint__id_seq";
CREATE SEQUENCE "public"."constraint__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 9
 CACHE 1;

-- ----------------------------
-- Sequence structure for constraint_args__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."constraint_args__id_seq";
CREATE SEQUENCE "public"."constraint_args__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 2
 CACHE 1;

-- ----------------------------
-- Sequence structure for constraint_dataman_field_type__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."constraint_dataman_field_type__id_seq";
CREATE SEQUENCE "public"."constraint_dataman_field_type__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 16
 CACHE 1;
SELECT setval('"public"."constraint_dataman_field_type__id_seq"', 16, true);

-- ----------------------------
-- Sequence structure for database__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database__id_seq";
CREATE SEQUENCE "public"."database__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1417
 CACHE 1;
SELECT setval('"public"."database__id_seq"', 1417, true);

-- ----------------------------
-- Sequence structure for database_datastore__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database_datastore__id_seq";
CREATE SEQUENCE "public"."database_datastore__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1351
 CACHE 1;
SELECT setval('"public"."database_datastore__id_seq"', 1351, true);

-- ----------------------------
-- Sequence structure for database_vshard__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database_vshard__id_seq";
CREATE SEQUENCE "public"."database_vshard__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1358
 CACHE 1;
SELECT setval('"public"."database_vshard__id_seq"', 1358, true);

-- ----------------------------
-- Sequence structure for database_vshard_instance__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database_vshard_instance__id_seq";
CREATE SEQUENCE "public"."database_vshard_instance__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 3379
 CACHE 1;
SELECT setval('"public"."database_vshard_instance__id_seq"', 3379, true);

-- ----------------------------
-- Sequence structure for database_vshard_instance_datastore_shard__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database_vshard_instance_datastore_shard__id_seq";
CREATE SEQUENCE "public"."database_vshard_instance_datastore_shard__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 20984
 CACHE 1;
SELECT setval('"public"."database_vshard_instance_datastore_shard__id_seq"', 20984, true);

-- ----------------------------
-- Sequence structure for dataman_field_type__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."dataman_field_type__id_seq";
CREATE SEQUENCE "public"."dataman_field_type__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 5
 CACHE 1;
SELECT setval('"public"."dataman_field_type__id_seq"', 5, true);

-- ----------------------------
-- Sequence structure for dataman_field_type_datasource_field_type__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."dataman_field_type_datasource_field_type__id_seq";
CREATE SEQUENCE "public"."dataman_field_type_datasource_field_type__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 16
 CACHE 1;
SELECT setval('"public"."dataman_field_type_datasource_field_type__id_seq"', 16, true);

-- ----------------------------
-- Sequence structure for datasource__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datasource__id_seq";
CREATE SEQUENCE "public"."datasource__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 2
 CACHE 1;
SELECT setval('"public"."datasource__id_seq"', 2, true);

-- ----------------------------
-- Sequence structure for datasource_field_type__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datasource_field_type__id_seq";
CREATE SEQUENCE "public"."datasource_field_type__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 16
 CACHE 1;
SELECT setval('"public"."datasource_field_type__id_seq"', 16, true);

-- ----------------------------
-- Sequence structure for datasource_field_type_arg__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datasource_field_type_arg__id_seq";
CREATE SEQUENCE "public"."datasource_field_type_arg__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;
SELECT setval('"public"."datasource_field_type_arg__id_seq"', 1, true);

-- ----------------------------
-- Sequence structure for datasource_instance__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datasource_instance__id_seq";
CREATE SEQUENCE "public"."datasource_instance__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 114
 CACHE 1;
SELECT setval('"public"."datasource_instance__id_seq"', 114, true);

-- ----------------------------
-- Sequence structure for datasource_instance_shard_instance__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datasource_instance_shard_instance__id_seq";
CREATE SEQUENCE "public"."datasource_instance_shard_instance__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 3211
 CACHE 1;
SELECT setval('"public"."datasource_instance_shard_instance__id_seq"', 3211, true);

-- ----------------------------
-- Sequence structure for datastore__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datastore__id_seq";
CREATE SEQUENCE "public"."datastore__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 54
 CACHE 1;
SELECT setval('"public"."datastore__id_seq"', 54, true);

-- ----------------------------
-- Sequence structure for datastore_shard__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datastore_shard__id_seq";
CREATE SEQUENCE "public"."datastore_shard__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 67
 CACHE 1;
SELECT setval('"public"."datastore_shard__id_seq"', 67, true);

-- ----------------------------
-- Sequence structure for datastore_shard_replica__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datastore_shard_replica__id_seq";
CREATE SEQUENCE "public"."datastore_shard_replica__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 66
 CACHE 1;
SELECT setval('"public"."datastore_shard_replica__id_seq"', 66, true);

-- ----------------------------
-- Sequence structure for field_type__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."field_type__id_seq";
CREATE SEQUENCE "public"."field_type__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 4
 CACHE 1;
SELECT setval('"public"."field_type__id_seq"', 4, true);

-- ----------------------------
-- Sequence structure for field_type_constraint__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."field_type_constraint__id_seq";
CREATE SEQUENCE "public"."field_type_constraint__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 2
 CACHE 1;

-- ----------------------------
-- Sequence structure for field_type_constraint_arg__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."field_type_constraint_arg__id_seq";
CREATE SEQUENCE "public"."field_type_constraint_arg__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 2
 CACHE 1;

-- ----------------------------
-- Sequence structure for field_type_datasource_field_type__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."field_type_datasource_field_type__id_seq";
CREATE SEQUENCE "public"."field_type_datasource_field_type__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 3
 CACHE 1;
SELECT setval('"public"."field_type_datasource_field_type__id_seq"', 3, true);

-- ----------------------------
-- Sequence structure for field_type_datasource_field_type_arg__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."field_type_datasource_field_type_arg__id_seq";
CREATE SEQUENCE "public"."field_type_datasource_field_type_arg__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;
SELECT setval('"public"."field_type_datasource_field_type_arg__id_seq"', 1, true);

-- ----------------------------
-- Sequence structure for field_type_datasource_type__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."field_type_datasource_type__id_seq";
CREATE SEQUENCE "public"."field_type_datasource_type__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for storage_node_type__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."storage_node_type__id_seq";
CREATE SEQUENCE "public"."storage_node_type__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 137
 CACHE 1;
SELECT setval('"public"."storage_node_type__id_seq"', 137, true);

-- ----------------------------
-- Table structure for collection
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection";
CREATE TABLE "public"."collection" (
"_id" int4 DEFAULT nextval('collection__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"database_id" int4,
"collection_vshard_id" int4,
"provision_state" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of collection
-- ----------------------------
INSERT INTO "public"."collection" VALUES ('4090', 'thread', '1417', null, '3');
INSERT INTO "public"."collection" VALUES ('4091', 'message', '1417', null, '3');
INSERT INTO "public"."collection" VALUES ('4092', 'user', '1417', null, '3');

-- ----------------------------
-- Table structure for collection_field
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_field";
CREATE TABLE "public"."collection_field" (
"_id" int4 DEFAULT nextval('collection_field__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"collection_id" int4,
"field_type" varchar(255) COLLATE "default",
"field_type_args" jsonb,
"not_null" bool NOT NULL,
"parent_collection_field_id" int4,
"provision_state" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of collection_field
-- ----------------------------
INSERT INTO "public"."collection_field" VALUES ('16999', 'data', '4090', 'document', 'null', 'f', null, '3');
INSERT INTO "public"."collection_field" VALUES ('17000', 'title', '4090', 'string', '{"size": 255}', 't', '16999', '3');
INSERT INTO "public"."collection_field" VALUES ('17001', 'created_by', '4090', 'string', '{"size": 255}', 't', '16999', '3');
INSERT INTO "public"."collection_field" VALUES ('17002', 'created', '4090', 'int', 'null', 't', '16999', '3');
INSERT INTO "public"."collection_field" VALUES ('17003', '_id', '4090', 'int', 'null', 't', null, '3');
INSERT INTO "public"."collection_field" VALUES ('17004', 'data', '4091', 'document', 'null', 'f', null, '3');
INSERT INTO "public"."collection_field" VALUES ('17005', 'created_by', '4091', 'string', '{"size": 255}', 't', '17004', '3');
INSERT INTO "public"."collection_field" VALUES ('17006', 'created', '4091', 'int', 'null', 't', '17004', '3');
INSERT INTO "public"."collection_field" VALUES ('17007', 'content', '4091', 'string', '{"size": 255}', 't', '17004', '3');
INSERT INTO "public"."collection_field" VALUES ('17008', 'thread_id', '4091', 'int', 'null', 't', '17004', '3');
INSERT INTO "public"."collection_field" VALUES ('17009', '_id', '4091', 'int', 'null', 't', null, '3');
INSERT INTO "public"."collection_field" VALUES ('17010', 'username', '4092', 'string', '{"size": 128}', 't', null, '3');
INSERT INTO "public"."collection_field" VALUES ('17011', '_id', '4092', 'int', 'null', 't', null, '3');

-- ----------------------------
-- Table structure for collection_field_relation
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_field_relation";
CREATE TABLE "public"."collection_field_relation" (
"_id" int4 DEFAULT nextval('collection_field_relation__id_seq'::regclass) NOT NULL,
"collection_field_id" int4 NOT NULL,
"relation_collection_field_id" int4 NOT NULL,
"cascade_on_delete" bool NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of collection_field_relation
-- ----------------------------
INSERT INTO "public"."collection_field_relation" VALUES ('1219', '17008', '17003', 'f');

-- ----------------------------
-- Table structure for collection_index
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_index";
CREATE TABLE "public"."collection_index" (
"_id" int4 DEFAULT nextval('collection_index__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"collection_id" int4,
"unique" bool,
"data_json" text COLLATE "default",
"provision_state" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of collection_index
-- ----------------------------
INSERT INTO "public"."collection_index" VALUES ('5317', 'created', '4090', 'f', null, '3');
INSERT INTO "public"."collection_index" VALUES ('5318', 'title', '4090', 't', null, '3');
INSERT INTO "public"."collection_index" VALUES ('5319', 'created', '4091', 'f', null, '3');
INSERT INTO "public"."collection_index" VALUES ('5320', 'username', '4092', 't', null, '3');

-- ----------------------------
-- Table structure for collection_index_item
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_index_item";
CREATE TABLE "public"."collection_index_item" (
"_id" int4 DEFAULT nextval('collection_index_item__id_seq'::regclass) NOT NULL,
"collection_index_id" int4 NOT NULL,
"collection_field_id" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of collection_index_item
-- ----------------------------
INSERT INTO "public"."collection_index_item" VALUES ('60324', '5317', '17002');
INSERT INTO "public"."collection_index_item" VALUES ('60325', '5318', '17000');
INSERT INTO "public"."collection_index_item" VALUES ('60326', '5319', '17006');
INSERT INTO "public"."collection_index_item" VALUES ('60327', '5320', '17010');

-- ----------------------------
-- Table structure for collection_partition
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_partition";
CREATE TABLE "public"."collection_partition" (
"_id" int4 DEFAULT nextval('collection_partition_id_seq'::regclass) NOT NULL,
"collection_id" int4 NOT NULL,
"start_id" int4 NOT NULL,
"end_id" int4,
"shard_config_json" jsonb
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of collection_partition
-- ----------------------------
INSERT INTO "public"."collection_partition" VALUES ('4086', '4090', '1', '0', '{"shard_key": "_id", "hash_method": "cast", "shard_method": "mod"}');
INSERT INTO "public"."collection_partition" VALUES ('4087', '4091', '1', '0', '{"shard_key": "_id", "hash_method": "cast", "shard_method": "mod"}');
INSERT INTO "public"."collection_partition" VALUES ('4088', '4092', '1', '0', '{"shard_key": "username", "hash_method": "sha256", "shard_method": "mod"}');

-- ----------------------------
-- Table structure for collection_vshard
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_vshard";
CREATE TABLE "public"."collection_vshard" (
"_id" int4 DEFAULT nextval('collection_vshard__id_seq'::regclass) NOT NULL,
"shard_count" int4
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of collection_vshard
-- ----------------------------

-- ----------------------------
-- Table structure for collection_vshard_instance
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_vshard_instance";
CREATE TABLE "public"."collection_vshard_instance" (
"_id" int4 DEFAULT nextval('collection_vshard_instance__id_seq'::regclass) NOT NULL,
"collection_vshard_id" int4,
"shard_instance" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of collection_vshard_instance
-- ----------------------------

-- ----------------------------
-- Table structure for collection_vshard_instance_datastore_shard
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_vshard_instance_datastore_shard";
CREATE TABLE "public"."collection_vshard_instance_datastore_shard" (
"_id" int4 DEFAULT nextval('collection_vshard_instance_datastore_shard__id_seq'::regclass) NOT NULL,
"collection_vshard_instance_id" int4,
"datastore_shard_id" int4
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of collection_vshard_instance_datastore_shard
-- ----------------------------

-- ----------------------------
-- Table structure for database
-- ----------------------------
DROP TABLE IF EXISTS "public"."database";
CREATE TABLE "public"."database" (
"_id" int4 DEFAULT nextval('database__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"provision_state" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of database
-- ----------------------------
INSERT INTO "public"."database" VALUES ('1417', 'example_forum', '5');

-- ----------------------------
-- Table structure for database_datastore
-- ----------------------------
DROP TABLE IF EXISTS "public"."database_datastore";
CREATE TABLE "public"."database_datastore" (
"_id" int4 DEFAULT nextval('database_datastore__id_seq'::regclass) NOT NULL,
"database_id" int4,
"datastore_id" int4,
"read" bool,
"write" bool,
"required" bool,
"provision_state" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of database_datastore
-- ----------------------------
INSERT INTO "public"."database_datastore" VALUES ('1351', '1417', '54', 't', 't', 't', '3');

-- ----------------------------
-- Table structure for database_vshard
-- ----------------------------
DROP TABLE IF EXISTS "public"."database_vshard";
CREATE TABLE "public"."database_vshard" (
"_id" int4 DEFAULT nextval('database_vshard__id_seq'::regclass) NOT NULL,
"shard_count" int4 NOT NULL,
"database_id" int4
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of database_vshard
-- ----------------------------
INSERT INTO "public"."database_vshard" VALUES ('1358', '2', '1417');

-- ----------------------------
-- Table structure for database_vshard_instance
-- ----------------------------
DROP TABLE IF EXISTS "public"."database_vshard_instance";
CREATE TABLE "public"."database_vshard_instance" (
"_id" int4 DEFAULT nextval('database_vshard_instance__id_seq'::regclass) NOT NULL,
"database_vshard_id" int4 NOT NULL,
"shard_instance" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of database_vshard_instance
-- ----------------------------
INSERT INTO "public"."database_vshard_instance" VALUES ('3378', '1358', '1');
INSERT INTO "public"."database_vshard_instance" VALUES ('3379', '1358', '2');

-- ----------------------------
-- Table structure for database_vshard_instance_datastore_shard
-- ----------------------------
DROP TABLE IF EXISTS "public"."database_vshard_instance_datastore_shard";
CREATE TABLE "public"."database_vshard_instance_datastore_shard" (
"_id" int4 DEFAULT nextval('database_vshard_instance_datastore_shard__id_seq'::regclass) NOT NULL,
"database_vshard_instance_id" int4,
"datastore_shard_id" int4
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of database_vshard_instance_datastore_shard
-- ----------------------------
INSERT INTO "public"."database_vshard_instance_datastore_shard" VALUES ('20975', '3378', '67');
INSERT INTO "public"."database_vshard_instance_datastore_shard" VALUES ('20976', '3379', '66');

-- ----------------------------
-- Table structure for datasource
-- ----------------------------
DROP TABLE IF EXISTS "public"."datasource";
CREATE TABLE "public"."datasource" (
"_id" int4 DEFAULT nextval('datasource__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default"
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of datasource
-- ----------------------------
INSERT INTO "public"."datasource" VALUES ('1', 'postgres cluster 1');
INSERT INTO "public"."datasource" VALUES ('2', 'cassandra');

-- ----------------------------
-- Table structure for datasource_instance
-- ----------------------------
DROP TABLE IF EXISTS "public"."datasource_instance";
CREATE TABLE "public"."datasource_instance" (
"_id" int4 DEFAULT nextval('datasource_instance__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"datasource_id" int4 NOT NULL,
"storage_node_id" int4 NOT NULL,
"config_json" jsonb,
"provision_state" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of datasource_instance
-- ----------------------------
INSERT INTO "public"."datasource_instance" VALUES ('113', 'postgres1', '1', '136', 'null', '3');
INSERT INTO "public"."datasource_instance" VALUES ('114', 'postgres1', '1', '137', 'null', '3');

-- ----------------------------
-- Table structure for datasource_instance_shard_instance
-- ----------------------------
DROP TABLE IF EXISTS "public"."datasource_instance_shard_instance";
CREATE TABLE "public"."datasource_instance_shard_instance" (
"_id" int4 DEFAULT nextval('datasource_instance_shard_instance__id_seq'::regclass) NOT NULL,
"datasource_instance_id" int4,
"database_vshard_instance_id" int4,
"collection_vshard_instance_id" int4,
"name" varchar(255) COLLATE "default",
"provision_state" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of datasource_instance_shard_instance
-- ----------------------------
INSERT INTO "public"."datasource_instance_shard_instance" VALUES ('3210', '113', '3378', null, 'dbshard_example_forum_1', '3');
INSERT INTO "public"."datasource_instance_shard_instance" VALUES ('3211', '114', '3379', null, 'dbshard_example_forum_2', '3');

-- ----------------------------
-- Table structure for datastore
-- ----------------------------
DROP TABLE IF EXISTS "public"."datastore";
CREATE TABLE "public"."datastore" (
"_id" int4 DEFAULT nextval('datastore__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"provision_state" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of datastore
-- ----------------------------
INSERT INTO "public"."datastore" VALUES ('54', 'test_datastore', '3');

-- ----------------------------
-- Table structure for datastore_shard
-- ----------------------------
DROP TABLE IF EXISTS "public"."datastore_shard";
CREATE TABLE "public"."datastore_shard" (
"_id" int4 DEFAULT nextval('datastore_shard__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"datastore_id" int4,
"shard_instance" int4 NOT NULL,
"provision_state" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of datastore_shard
-- ----------------------------
INSERT INTO "public"."datastore_shard" VALUES ('66', 'datastore_test-shard1', '54', '1', '3');
INSERT INTO "public"."datastore_shard" VALUES ('67', 'test-shard2', '54', '2', '3');

-- ----------------------------
-- Table structure for datastore_shard_replica
-- ----------------------------
DROP TABLE IF EXISTS "public"."datastore_shard_replica";
CREATE TABLE "public"."datastore_shard_replica" (
"_id" int4 DEFAULT nextval('datastore_shard_replica__id_seq'::regclass) NOT NULL,
"datastore_shard_id" int4,
"datasource_instance_id" int4,
"master" bool NOT NULL,
"provision_state" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of datastore_shard_replica
-- ----------------------------
INSERT INTO "public"."datastore_shard_replica" VALUES ('65', '66', '114', 't', '3');
INSERT INTO "public"."datastore_shard_replica" VALUES ('66', '67', '113', 't', '3');

-- ----------------------------
-- Table structure for field_type
-- ----------------------------
DROP TABLE IF EXISTS "public"."field_type";
CREATE TABLE "public"."field_type" (
"_id" int4 DEFAULT nextval('field_type__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"dataman_type" varchar(255) COLLATE "default"
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of field_type
-- ----------------------------
INSERT INTO "public"."field_type" VALUES ('1', 'age', 'int');
INSERT INTO "public"."field_type" VALUES ('2', 'phone number', 'string');
INSERT INTO "public"."field_type" VALUES ('3', 'string', 'string');
INSERT INTO "public"."field_type" VALUES ('4', 'int', 'int');

-- ----------------------------
-- Table structure for field_type_constraint
-- ----------------------------
DROP TABLE IF EXISTS "public"."field_type_constraint";
CREATE TABLE "public"."field_type_constraint" (
"_id" int4 DEFAULT nextval('field_type_constraint__id_seq'::regclass) NOT NULL,
"field_type_id" int4 NOT NULL,
"constraint" varchar(255) COLLATE "default" NOT NULL,
"args" jsonb
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of field_type_constraint
-- ----------------------------
INSERT INTO "public"."field_type_constraint" VALUES ('1', '1', '<', '{"value": 200}');

-- ----------------------------
-- Table structure for field_type_datasource_type
-- ----------------------------
DROP TABLE IF EXISTS "public"."field_type_datasource_type";
CREATE TABLE "public"."field_type_datasource_type" (
"_id" int4 DEFAULT nextval('field_type_datasource_type__id_seq'::regclass) NOT NULL,
"field_type_id" int4 NOT NULL,
"datasource_type" varchar(255) COLLATE "default" NOT NULL,
"args" jsonb
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of field_type_datasource_type
-- ----------------------------

-- ----------------------------
-- Table structure for storage_node
-- ----------------------------
DROP TABLE IF EXISTS "public"."storage_node";
CREATE TABLE "public"."storage_node" (
"_id" int4 DEFAULT nextval('storage_node_type__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"ip" varchar(255) COLLATE "default",
"port" int4,
"provision_state" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of storage_node
-- ----------------------------
INSERT INTO "public"."storage_node" VALUES ('136', 'NUC', '10.42.17.93', '8081', '3');
INSERT INTO "public"."storage_node" VALUES ('137', 'X1', '127.0.0.1', '8081', '3');

-- ----------------------------
-- Alter Sequences Owned By 
-- ----------------------------
ALTER SEQUENCE "public"."collection__id_seq" OWNED BY "collection"."_id";
ALTER SEQUENCE "public"."collection_field__id_seq" OWNED BY "collection_field"."_id";
ALTER SEQUENCE "public"."collection_field_relation__id_seq" OWNED BY "collection_field_relation"."_id";
ALTER SEQUENCE "public"."collection_index__id_seq" OWNED BY "collection_index"."_id";
ALTER SEQUENCE "public"."collection_index_item__id_seq" OWNED BY "collection_index_item"."_id";
ALTER SEQUENCE "public"."collection_partition_id_seq" OWNED BY "collection_partition"."_id";
ALTER SEQUENCE "public"."collection_vshard__id_seq" OWNED BY "collection_vshard"."_id";
ALTER SEQUENCE "public"."collection_vshard_instance__id_seq" OWNED BY "collection_vshard_instance"."_id";
ALTER SEQUENCE "public"."collection_vshard_instance_datastore_shard__id_seq" OWNED BY "collection_vshard_instance_datastore_shard"."_id";
ALTER SEQUENCE "public"."database__id_seq" OWNED BY "database"."_id";
ALTER SEQUENCE "public"."database_datastore__id_seq" OWNED BY "database_datastore"."_id";
ALTER SEQUENCE "public"."database_vshard__id_seq" OWNED BY "database_vshard"."_id";
ALTER SEQUENCE "public"."database_vshard_instance__id_seq" OWNED BY "database_vshard_instance"."_id";
ALTER SEQUENCE "public"."database_vshard_instance_datastore_shard__id_seq" OWNED BY "database_vshard_instance_datastore_shard"."_id";
ALTER SEQUENCE "public"."datasource__id_seq" OWNED BY "datasource"."_id";
ALTER SEQUENCE "public"."datasource_instance__id_seq" OWNED BY "datasource_instance"."_id";
ALTER SEQUENCE "public"."datasource_instance_shard_instance__id_seq" OWNED BY "datasource_instance_shard_instance"."_id";
ALTER SEQUENCE "public"."datastore__id_seq" OWNED BY "datastore"."_id";
ALTER SEQUENCE "public"."datastore_shard__id_seq" OWNED BY "datastore_shard"."_id";
ALTER SEQUENCE "public"."datastore_shard_replica__id_seq" OWNED BY "datastore_shard_replica"."_id";
ALTER SEQUENCE "public"."field_type_datasource_type__id_seq" OWNED BY "field_type_datasource_type"."_id";
ALTER SEQUENCE "public"."storage_node_type__id_seq" OWNED BY "storage_node"."_id";

-- ----------------------------
-- Indexes structure for table collection
-- ----------------------------
CREATE UNIQUE INDEX "index_index_collection_collection_name" ON "public"."collection" USING btree ("name", "database_id");

-- ----------------------------
-- Primary Key structure for table collection
-- ----------------------------
ALTER TABLE "public"."collection" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_field
-- ----------------------------
CREATE UNIQUE INDEX "collection_field_name_collection_id_parent_collection_field_idx" ON "public"."collection_field" USING btree ("name", "collection_id", "parent_collection_field_id");

-- ----------------------------
-- Primary Key structure for table collection_field
-- ----------------------------
ALTER TABLE "public"."collection_field" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_field_relation
-- ----------------------------
CREATE UNIQUE INDEX "collection_field_relation_collection_field_id_idx" ON "public"."collection_field_relation" USING btree ("collection_field_id");

-- ----------------------------
-- Primary Key structure for table collection_field_relation
-- ----------------------------
ALTER TABLE "public"."collection_field_relation" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_index
-- ----------------------------
CREATE UNIQUE INDEX "index_collection_index_name" ON "public"."collection_index" USING btree ("name", "collection_id");

-- ----------------------------
-- Primary Key structure for table collection_index
-- ----------------------------
ALTER TABLE "public"."collection_index" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_index_item
-- ----------------------------
CREATE UNIQUE INDEX "collection_index_item_collection_index_id_collection_field__idx" ON "public"."collection_index_item" USING btree ("collection_index_id", "collection_field_id");

-- ----------------------------
-- Primary Key structure for table collection_index_item
-- ----------------------------
ALTER TABLE "public"."collection_index_item" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_partition
-- ----------------------------
CREATE INDEX "collection_partition_collection_id_idx" ON "public"."collection_partition" USING btree ("collection_id");
CREATE INDEX "toremove" ON "public"."collection_partition" USING btree ("collection_id");

-- ----------------------------
-- Primary Key structure for table collection_partition
-- ----------------------------
ALTER TABLE "public"."collection_partition" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Primary Key structure for table collection_vshard
-- ----------------------------
ALTER TABLE "public"."collection_vshard" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_vshard_instance
-- ----------------------------
CREATE UNIQUE INDEX "collection_vshard_instance_collection_vshard_id_shard_insta_idx" ON "public"."collection_vshard_instance" USING btree ("collection_vshard_id", "shard_instance");

-- ----------------------------
-- Primary Key structure for table collection_vshard_instance
-- ----------------------------
ALTER TABLE "public"."collection_vshard_instance" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_vshard_instance_datastore_shard
-- ----------------------------
CREATE UNIQUE INDEX "collection_vshard_instance_da_collection_vshard_instance_id_idx" ON "public"."collection_vshard_instance_datastore_shard" USING btree ("collection_vshard_instance_id");

-- ----------------------------
-- Primary Key structure for table collection_vshard_instance_datastore_shard
-- ----------------------------
ALTER TABLE "public"."collection_vshard_instance_datastore_shard" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table database
-- ----------------------------
CREATE UNIQUE INDEX "index_index_database_name" ON "public"."database" USING btree ("name");

-- ----------------------------
-- Primary Key structure for table database
-- ----------------------------
ALTER TABLE "public"."database" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table database_datastore
-- ----------------------------
CREATE UNIQUE INDEX "database_datastore_database_id_datastore_id_idx" ON "public"."database_datastore" USING btree ("database_id", "datastore_id");
CREATE INDEX "database_id_idx" ON "public"."database_datastore" USING btree ("database_id");

-- ----------------------------
-- Primary Key structure for table database_datastore
-- ----------------------------
ALTER TABLE "public"."database_datastore" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table database_vshard
-- ----------------------------
CREATE UNIQUE INDEX "database_vshard_database_id_idx" ON "public"."database_vshard" USING btree ("database_id");

-- ----------------------------
-- Primary Key structure for table database_vshard
-- ----------------------------
ALTER TABLE "public"."database_vshard" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table database_vshard_instance
-- ----------------------------
CREATE UNIQUE INDEX "database_vshard_instance_database_vshard_id_shard_instance_idx" ON "public"."database_vshard_instance" USING btree ("database_vshard_id", "shard_instance");

-- ----------------------------
-- Primary Key structure for table database_vshard_instance
-- ----------------------------
ALTER TABLE "public"."database_vshard_instance" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table database_vshard_instance_datastore_shard
-- ----------------------------
CREATE UNIQUE INDEX "database_vshard_instance_datast_database_vshard_instance_id_idx" ON "public"."database_vshard_instance_datastore_shard" USING btree ("database_vshard_instance_id");

-- ----------------------------
-- Primary Key structure for table database_vshard_instance_datastore_shard
-- ----------------------------
ALTER TABLE "public"."database_vshard_instance_datastore_shard" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table datasource
-- ----------------------------
CREATE UNIQUE INDEX "datasource_name_idx" ON "public"."datasource" USING btree ("name");

-- ----------------------------
-- Primary Key structure for table datasource
-- ----------------------------
ALTER TABLE "public"."datasource" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table datasource_instance
-- ----------------------------
CREATE UNIQUE INDEX "datasource_instance_name_storage_node_id_idx" ON "public"."datasource_instance" USING btree ("name", "storage_node_id");

-- ----------------------------
-- Primary Key structure for table datasource_instance
-- ----------------------------
ALTER TABLE "public"."datasource_instance" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table datasource_instance_shard_instance
-- ----------------------------
CREATE UNIQUE INDEX "datasource_instance_shard_ins_datasource_instance_id_databa_idx" ON "public"."datasource_instance_shard_instance" USING btree ("datasource_instance_id", "database_vshard_instance_id", "collection_vshard_instance_id");
CREATE UNIQUE INDEX "datasource_instance_shard_insta_datasource_instance_id_name_idx" ON "public"."datasource_instance_shard_instance" USING btree ("datasource_instance_id", "name");

-- ----------------------------
-- Primary Key structure for table datasource_instance_shard_instance
-- ----------------------------
ALTER TABLE "public"."datasource_instance_shard_instance" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table datastore
-- ----------------------------
CREATE UNIQUE INDEX "datastore_name_idx" ON "public"."datastore" USING btree ("name");

-- ----------------------------
-- Primary Key structure for table datastore
-- ----------------------------
ALTER TABLE "public"."datastore" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table datastore_shard
-- ----------------------------
CREATE UNIQUE INDEX "datastore_shard_name_datastore_id_idx" ON "public"."datastore_shard" USING btree ("name", "datastore_id");
CREATE UNIQUE INDEX "datastore_shard_number" ON "public"."datastore_shard" USING btree ("datastore_id", "shard_instance");

-- ----------------------------
-- Primary Key structure for table datastore_shard
-- ----------------------------
ALTER TABLE "public"."datastore_shard" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table datastore_shard_replica
-- ----------------------------
CREATE UNIQUE INDEX "datastore_shard_replica_datastore_shard_id_datasource_insta_idx" ON "public"."datastore_shard_replica" USING btree ("datastore_shard_id", "datasource_instance_id");

-- ----------------------------
-- Primary Key structure for table datastore_shard_replica
-- ----------------------------
ALTER TABLE "public"."datastore_shard_replica" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table field_type
-- ----------------------------
CREATE UNIQUE INDEX "field_type_name_idx" ON "public"."field_type" USING btree ("name");

-- ----------------------------
-- Primary Key structure for table field_type
-- ----------------------------
ALTER TABLE "public"."field_type" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table field_type_constraint
-- ----------------------------
CREATE INDEX "field_type_constraint_field_type_id_constraint_id_idx" ON "public"."field_type_constraint" USING btree ("field_type_id", "constraint");

-- ----------------------------
-- Primary Key structure for table field_type_constraint
-- ----------------------------
ALTER TABLE "public"."field_type_constraint" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table field_type_datasource_type
-- ----------------------------
CREATE UNIQUE INDEX "field_type_datasource_type_field_type_id_datasource_type_idx" ON "public"."field_type_datasource_type" USING btree ("field_type_id", "datasource_type");

-- ----------------------------
-- Primary Key structure for table field_type_datasource_type
-- ----------------------------
ALTER TABLE "public"."field_type_datasource_type" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table storage_node
-- ----------------------------
CREATE UNIQUE INDEX "storage_node_ip_port_idx" ON "public"."storage_node" USING btree ("ip", "port");
CREATE UNIQUE INDEX "storage_node_name_idx" ON "public"."storage_node" USING btree ("name");

-- ----------------------------
-- Primary Key structure for table storage_node
-- ----------------------------
ALTER TABLE "public"."storage_node" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Foreign Key structure for table "public"."collection"
-- ----------------------------
ALTER TABLE "public"."collection" ADD FOREIGN KEY ("collection_vshard_id") REFERENCES "public"."collection_vshard" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."collection" ADD FOREIGN KEY ("database_id") REFERENCES "public"."database" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_field"
-- ----------------------------
ALTER TABLE "public"."collection_field" ADD FOREIGN KEY ("parent_collection_field_id") REFERENCES "public"."collection_field" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."collection_field" ADD FOREIGN KEY ("collection_id") REFERENCES "public"."collection" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_field_relation"
-- ----------------------------
ALTER TABLE "public"."collection_field_relation" ADD FOREIGN KEY ("collection_field_id") REFERENCES "public"."collection_field" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."collection_field_relation" ADD FOREIGN KEY ("relation_collection_field_id") REFERENCES "public"."collection_field" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_index"
-- ----------------------------
ALTER TABLE "public"."collection_index" ADD FOREIGN KEY ("collection_id") REFERENCES "public"."collection" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_index_item"
-- ----------------------------
ALTER TABLE "public"."collection_index_item" ADD FOREIGN KEY ("collection_field_id") REFERENCES "public"."collection_field" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."collection_index_item" ADD FOREIGN KEY ("collection_index_id") REFERENCES "public"."collection_index" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_partition"
-- ----------------------------
ALTER TABLE "public"."collection_partition" ADD FOREIGN KEY ("collection_id") REFERENCES "public"."collection" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_vshard_instance"
-- ----------------------------
ALTER TABLE "public"."collection_vshard_instance" ADD FOREIGN KEY ("collection_vshard_id") REFERENCES "public"."collection_vshard" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_vshard_instance_datastore_shard"
-- ----------------------------
ALTER TABLE "public"."collection_vshard_instance_datastore_shard" ADD FOREIGN KEY ("collection_vshard_instance_id") REFERENCES "public"."collection_vshard_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."collection_vshard_instance_datastore_shard" ADD FOREIGN KEY ("datastore_shard_id") REFERENCES "public"."datastore_shard" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."database_datastore"
-- ----------------------------
ALTER TABLE "public"."database_datastore" ADD FOREIGN KEY ("database_id") REFERENCES "public"."database" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."database_datastore" ADD FOREIGN KEY ("datastore_id") REFERENCES "public"."datastore" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."database_vshard"
-- ----------------------------
ALTER TABLE "public"."database_vshard" ADD FOREIGN KEY ("database_id") REFERENCES "public"."database" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."database_vshard_instance"
-- ----------------------------
ALTER TABLE "public"."database_vshard_instance" ADD FOREIGN KEY ("database_vshard_id") REFERENCES "public"."database_vshard" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."database_vshard_instance_datastore_shard"
-- ----------------------------
ALTER TABLE "public"."database_vshard_instance_datastore_shard" ADD FOREIGN KEY ("database_vshard_instance_id") REFERENCES "public"."database_vshard_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."database_vshard_instance_datastore_shard" ADD FOREIGN KEY ("datastore_shard_id") REFERENCES "public"."datastore_shard" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datasource_instance"
-- ----------------------------
ALTER TABLE "public"."datasource_instance" ADD FOREIGN KEY ("storage_node_id") REFERENCES "public"."storage_node" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."datasource_instance" ADD FOREIGN KEY ("datasource_id") REFERENCES "public"."datasource" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datasource_instance_shard_instance"
-- ----------------------------
ALTER TABLE "public"."datasource_instance_shard_instance" ADD FOREIGN KEY ("collection_vshard_instance_id") REFERENCES "public"."collection_vshard_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."datasource_instance_shard_instance" ADD FOREIGN KEY ("database_vshard_instance_id") REFERENCES "public"."database_vshard_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."datasource_instance_shard_instance" ADD FOREIGN KEY ("datasource_instance_id") REFERENCES "public"."datasource_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datastore_shard"
-- ----------------------------
ALTER TABLE "public"."datastore_shard" ADD FOREIGN KEY ("datastore_id") REFERENCES "public"."datastore" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datastore_shard_replica"
-- ----------------------------
ALTER TABLE "public"."datastore_shard_replica" ADD FOREIGN KEY ("datastore_shard_id") REFERENCES "public"."datastore_shard" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."datastore_shard_replica" ADD FOREIGN KEY ("datasource_instance_id") REFERENCES "public"."datasource_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
