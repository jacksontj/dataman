/*
Navicat PGSQL Data Transfer

Source Server         : local
Source Server Version : 90608
Source Host           : localhost:5432
Source Database       : dataman_storage
Source Schema         : public

Target Server Type    : PGSQL
Target Server Version : 90608
File Encoding         : 65001

Date: 2018-03-23 21:33:36
*/


-- ----------------------------
-- Sequence structure for collection__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection__id_seq";
CREATE SEQUENCE "public"."collection__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 7265
 CACHE 1;
SELECT setval('"public"."collection__id_seq"', 7265, true);

-- ----------------------------
-- Sequence structure for collection_field__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_field__id_seq";
CREATE SEQUENCE "public"."collection_field__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 26701
 CACHE 1;
SELECT setval('"public"."collection_field__id_seq"', 26701, true);

-- ----------------------------
-- Sequence structure for collection_field_relation__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_field_relation__id_seq";
CREATE SEQUENCE "public"."collection_field_relation__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 2465
 CACHE 1;
SELECT setval('"public"."collection_field_relation__id_seq"', 2465, true);

-- ----------------------------
-- Sequence structure for collection_index__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_index__id_seq";
CREATE SEQUENCE "public"."collection_index__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 9009
 CACHE 1;
SELECT setval('"public"."collection_index__id_seq"', 9009, true);

-- ----------------------------
-- Sequence structure for collection_index_item__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_index_item__id_seq";
CREATE SEQUENCE "public"."collection_index_item__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 144620
 CACHE 1;
SELECT setval('"public"."collection_index_item__id_seq"', 144620, true);

-- ----------------------------
-- Sequence structure for database__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database__id_seq";
CREATE SEQUENCE "public"."database__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 2452
 CACHE 1;
SELECT setval('"public"."database__id_seq"', 2452, true);

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
-- Sequence structure for shard_instance__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."shard_instance__id_seq";
CREATE SEQUENCE "public"."shard_instance__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 3133
 CACHE 1;
SELECT setval('"public"."shard_instance__id_seq"', 3133, true);

-- ----------------------------
-- Table structure for collection
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection";
CREATE TABLE "public"."collection" (
"_id" int4 DEFAULT nextval('collection__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"shard_instance_id" int4 NOT NULL,
"provision_state" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of collection
-- ----------------------------
INSERT INTO "public"."collection" VALUES ('4969', 'thread', '1721', '0');
INSERT INTO "public"."collection" VALUES ('4970', 'message', '1721', '0');
INSERT INTO "public"."collection" VALUES ('4971', 'user', '1721', '0');
INSERT INTO "public"."collection" VALUES ('5631', 'database', '2158', '3');
INSERT INTO "public"."collection" VALUES ('5632', 'datastore', '2158', '3');
INSERT INTO "public"."collection" VALUES ('5633', 'datastore_vshard', '2158', '3');
INSERT INTO "public"."collection" VALUES ('5634', 'collection', '2158', '3');
INSERT INTO "public"."collection" VALUES ('5635', 'collection_keyspace', '2158', '3');
INSERT INTO "public"."collection" VALUES ('5636', 'collection_keyspace_partition', '2158', '3');
INSERT INTO "public"."collection" VALUES ('5637', 'datasource', '2158', '3');
INSERT INTO "public"."collection" VALUES ('5638', 'collection_index', '2158', '3');
INSERT INTO "public"."collection" VALUES ('5639', 'database', '2159', '3');
INSERT INTO "public"."collection" VALUES ('5640', 'shard_instance', '2159', '3');
INSERT INTO "public"."collection" VALUES ('5641', 'collection', '2159', '3');
INSERT INTO "public"."collection" VALUES ('5642', 'collection_index', '2159', '3');
INSERT INTO "public"."collection" VALUES ('5643', 'collection_field', '2159', '3');
INSERT INTO "public"."collection" VALUES ('5644', 'field_type', '2159', '3');
INSERT INTO "public"."collection" VALUES ('5645', 'field_type_constraint', '2159', '3');
INSERT INTO "public"."collection" VALUES ('5646', 'collection_field_relation', '2159', '3');
INSERT INTO "public"."collection" VALUES ('5647', 'collection_index_item', '2159', '3');

-- ----------------------------
-- Table structure for collection_field
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_field";
CREATE TABLE "public"."collection_field" (
"_id" int4 DEFAULT nextval('collection_field__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"collection_id" int4,
"field_type" varchar(255) COLLATE "default",
"parent_collection_field_id" int4,
"provision_state" int4 NOT NULL,
"not_null" bool NOT NULL,
"default" varchar(255) COLLATE "default"
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of collection_field
-- ----------------------------
INSERT INTO "public"."collection_field" VALUES ('19997', 'id', '4969', '_int', '0', '0', 't', null);
INSERT INTO "public"."collection_field" VALUES ('19998', 'data', '4969', '_document', '0', '0', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('19999', 'created', '4969', '_int', '19998', '0', 't', null);
INSERT INTO "public"."collection_field" VALUES ('20000', 'created_by', '4969', '_string', '19998', '0', 't', null);
INSERT INTO "public"."collection_field" VALUES ('20001', 'title', '4969', '_string', '19998', '0', 't', null);
INSERT INTO "public"."collection_field" VALUES ('20002', 'id', '4970', '_int', '0', '0', 't', null);
INSERT INTO "public"."collection_field" VALUES ('20003', 'data', '4970', '_document', '0', '0', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('20004', 'thread_id', '4970', '_int', '20003', '0', 't', null);
INSERT INTO "public"."collection_field" VALUES ('20005', 'content', '4970', '_string', '20003', '0', 't', null);
INSERT INTO "public"."collection_field" VALUES ('20006', 'created', '4970', '_int', '20003', '0', 't', null);
INSERT INTO "public"."collection_field" VALUES ('20007', 'created_by', '4970', '_string', '20003', '0', 't', null);
INSERT INTO "public"."collection_field" VALUES ('20008', 'username', '4971', '_string', '0', '0', 't', null);
INSERT INTO "public"."collection_field" VALUES ('20009', 'id', '4971', '_int', '0', '0', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21914', 'name', '5631', '_string', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21915', 'provision_state', '5631', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21916', '_id', '5631', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21917', 'provision_state', '5632', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21918', '_id', '5632', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21919', 'name', '5632', '_string', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21920', 'datastore_id', '5633', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21921', 'name', '5633', '_string', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21922', 'shard_count', '5633', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21923', '_id', '5633', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21924', 'database_id', '5633', '_int', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21925', '_id', '5634', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21926', 'database_id', '5634', '_int', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21927', 'name', '5634', '_string', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21928', 'provision_state', '5634', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21929', '_id', '5635', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21930', 'collection_id', '5635', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21931', 'hash_method', '5635', '_string', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21932', 'parent_collection_keyspace_partition_id', '5635', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21933', '_id', '5636', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21934', 'collection_keyspace_id', '5636', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21935', 'end_id', '5636', '_int', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21936', 'shard_method', '5636', '_string', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21937', 'start_id', '5636', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21938', '_id', '5637', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21939', 'name', '5637', '_string', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21940', 'provision_state', '5638', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21941', 'unique', '5638', '_bool', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21942', '_id', '5638', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21943', 'collection_id', '5638', '_int', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21944', 'name', '5638', '_string', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21945', 'primary', '5638', '_bool', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21946', '_id', '5639', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21947', 'name', '5639', '_string', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21948', 'provision_state', '5639', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21949', 'name', '5640', '_string', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21950', 'provision_state', '5640', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21951', '_id', '5640', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21952', 'collection_shard', '5640', '_bool', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21953', 'count', '5640', '_int', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21954', 'database_id', '5640', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21955', 'database_shard', '5640', '_bool', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21956', 'instance', '5640', '_int', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21957', '_id', '5641', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21958', 'name', '5641', '_string', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21959', 'provision_state', '5641', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21960', 'shard_instance_id', '5641', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21961', 'provision_state', '5642', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21962', 'unique', '5642', '_bool', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21963', '_id', '5642', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21964', 'collection_id', '5642', '_int', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21965', 'name', '5642', '_string', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21966', 'primary', '5642', '_bool', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21967', 'collection_id', '5643', '_int', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21968', 'default', '5643', '_string', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21969', 'field_type', '5643', '_string', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21970', 'name', '5643', '_string', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21971', 'not_null', '5643', '_bool', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21972', 'parent_collection_field_id', '5643', '_int', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21973', 'provision_state', '5643', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21974', '_id', '5643', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21975', '_id', '5644', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21976', 'dataman_type', '5644', '_string', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21977', 'name', '5644', '_string', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21978', 'field_type_id', '5645', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21979', 'validation_error', '5645', '_string', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21980', 'args', '5645', '_json', '0', '3', 'f', null);
INSERT INTO "public"."collection_field" VALUES ('21981', 'constraint', '5645', '_string', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21982', 'f', '5645', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21983', 'relation_collection_field_id', '5646', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21984', '_id', '5646', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21985', 'cascade_on_delete', '5646', '_bool', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21986', 'collection_field_id', '5646', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21987', 'foreign_key', '5646', '_bool', '0', '3', 't', 'false');
INSERT INTO "public"."collection_field" VALUES ('21988', '_id', '5647', '_serial', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21989', 'collection_field_id', '5647', '_int', '0', '3', 't', null);
INSERT INTO "public"."collection_field" VALUES ('21990', 'collection_index_id', '5647', '_int', '0', '3', 't', null);

-- ----------------------------
-- Table structure for collection_field_relation
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_field_relation";
CREATE TABLE "public"."collection_field_relation" (
"_id" int4 DEFAULT nextval('collection_field_relation__id_seq'::regclass) NOT NULL,
"collection_field_id" int4 NOT NULL,
"relation_collection_field_id" int4 NOT NULL,
"cascade_on_delete" bool NOT NULL,
"foreign_key" bool DEFAULT false NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of collection_field_relation
-- ----------------------------
INSERT INTO "public"."collection_field_relation" VALUES ('1479', '20004', '19997', 'f', 'f');
INSERT INTO "public"."collection_field_relation" VALUES ('1743', '21920', '21918', 'f', 't');
INSERT INTO "public"."collection_field_relation" VALUES ('1744', '21924', '21916', 'f', 't');
INSERT INTO "public"."collection_field_relation" VALUES ('1745', '21926', '21916', 'f', 't');
INSERT INTO "public"."collection_field_relation" VALUES ('1746', '21930', '21925', 'f', 't');
INSERT INTO "public"."collection_field_relation" VALUES ('1747', '21934', '21929', 'f', 't');
INSERT INTO "public"."collection_field_relation" VALUES ('1748', '21943', '21925', 'f', 't');
INSERT INTO "public"."collection_field_relation" VALUES ('1749', '21954', '21946', 'f', 't');
INSERT INTO "public"."collection_field_relation" VALUES ('1750', '21960', '21951', 'f', 't');
INSERT INTO "public"."collection_field_relation" VALUES ('1751', '21964', '21957', 'f', 't');
INSERT INTO "public"."collection_field_relation" VALUES ('1752', '21967', '21957', 'f', 't');
INSERT INTO "public"."collection_field_relation" VALUES ('1753', '21983', '21974', 'f', 't');
INSERT INTO "public"."collection_field_relation" VALUES ('1754', '21986', '21974', 'f', 't');
INSERT INTO "public"."collection_field_relation" VALUES ('1755', '21989', '21974', 'f', 't');
INSERT INTO "public"."collection_field_relation" VALUES ('1756', '21990', '21963', 'f', 't');

-- ----------------------------
-- Table structure for collection_index
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_index";
CREATE TABLE "public"."collection_index" (
"_id" int4 DEFAULT nextval('collection_index__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"collection_id" int4,
"unique" bool,
"provision_state" int4 NOT NULL,
"primary" bool
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of collection_index
-- ----------------------------
INSERT INTO "public"."collection_index" VALUES ('6370', 'id', '4969', 't', '0', 't');
INSERT INTO "public"."collection_index" VALUES ('6371', 'created', '4969', 'f', '0', null);
INSERT INTO "public"."collection_index" VALUES ('6372', 'title', '4969', 't', '0', null);
INSERT INTO "public"."collection_index" VALUES ('6373', 'id', '4970', 't', '0', 't');
INSERT INTO "public"."collection_index" VALUES ('6374', 'created', '4970', 'f', '0', null);
INSERT INTO "public"."collection_index" VALUES ('6375', 'id', '4971', 't', '0', 't');
INSERT INTO "public"."collection_index" VALUES ('6376', 'username', '4971', 't', '0', null);
INSERT INTO "public"."collection_index" VALUES ('7163', 'database_pkey', '5631', 't', '3', 't');
INSERT INTO "public"."collection_index" VALUES ('7164', 'index_index_database_name', '5631', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7165', 'datastore_name_idx', '5632', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7166', 'datastore_pkey', '5632', 't', '3', 't');
INSERT INTO "public"."collection_index" VALUES ('7167', 'datastore_vshard_datastore_id_name_idx', '5633', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7168', 'datastore_vshard_pkey', '5633', 't', '3', 't');
INSERT INTO "public"."collection_index" VALUES ('7169', 'collection_pkey', '5634', 't', '3', 't');
INSERT INTO "public"."collection_index" VALUES ('7170', 'index_index_collection_collection_name', '5634', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7171', 'collection_keyspace_TOREMOVE', '5635', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7172', 'collection_keyspace_collection_id_parent_collection_keyspac_idx', '5635', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7173', 'collection_keyspace_pkey', '5635', 't', '3', 't');
INSERT INTO "public"."collection_index" VALUES ('7174', 'collection_keyspace_partition_TOREMOVE', '5636', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7175', 'collection_keyspace_partition_pkey', '5636', 't', '3', 't');
INSERT INTO "public"."collection_index" VALUES ('7176', 'datasource_name_idx', '5637', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7177', 'datasource_pkey', '5637', 't', '3', 't');
INSERT INTO "public"."collection_index" VALUES ('7178', 'database_name_idx', '5639', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7179', 'database_pkey', '5639', 't', '3', 't');
INSERT INTO "public"."collection_index" VALUES ('7180', 'shard_instance_name_database_id_idx', '5640', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7181', 'shard_instance_pkey', '5640', 't', '3', 't');
INSERT INTO "public"."collection_index" VALUES ('7182', 'shard_instance_database_id_count_instance_database_shard_co_idx', '5640', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7183', 'collection_name_shard_instance_id_idx', '5641', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7184', 'collection_pkey', '5641', 't', '3', 't');
INSERT INTO "public"."collection_index" VALUES ('7185', 'collection_field_pkey', '5643', 't', '3', 't');
INSERT INTO "public"."collection_index" VALUES ('7186', 'index_collection_field_collection_field_name', '5643', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7187', 'field_type_name_idx', '5644', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7188', 'field_type_pkey', '5644', 't', '3', 't');
INSERT INTO "public"."collection_index" VALUES ('7189', 'collection_field_relation_pkey', '5646', 't', '3', 't');
INSERT INTO "public"."collection_index" VALUES ('7190', 'collection_index_collection_id_primary_idx', '5642', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7191', 'collection_index_name', '5642', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7192', 'collection_index_pkey', '5642', 't', '3', 't');
INSERT INTO "public"."collection_index" VALUES ('7193', 'collection_index_item_collection_index_id_collection_field__idx', '5647', 't', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7194', 'collection_index_item_pkey', '5647', 't', '3', 't');
INSERT INTO "public"."collection_index" VALUES ('7195', 'field_type_constraint_field_type_id_constraint_id_idx', '5645', 'f', '3', null);
INSERT INTO "public"."collection_index" VALUES ('7196', 'field_type_constraint_pkey', '5645', 't', '3', 't');

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
INSERT INTO "public"."collection_index_item" VALUES ('98136', '6370', '19997');
INSERT INTO "public"."collection_index_item" VALUES ('98137', '6371', '19999');
INSERT INTO "public"."collection_index_item" VALUES ('98138', '6372', '20001');
INSERT INTO "public"."collection_index_item" VALUES ('98139', '6373', '20002');
INSERT INTO "public"."collection_index_item" VALUES ('98140', '6374', '20006');
INSERT INTO "public"."collection_index_item" VALUES ('98144', '6375', '20009');
INSERT INTO "public"."collection_index_item" VALUES ('98145', '6376', '20008');
INSERT INTO "public"."collection_index_item" VALUES ('112220', '7163', '21916');
INSERT INTO "public"."collection_index_item" VALUES ('112221', '7164', '21914');
INSERT INTO "public"."collection_index_item" VALUES ('112222', '7165', '21919');
INSERT INTO "public"."collection_index_item" VALUES ('112223', '7166', '21918');
INSERT INTO "public"."collection_index_item" VALUES ('112224', '7167', '21920');
INSERT INTO "public"."collection_index_item" VALUES ('112225', '7167', '21921');
INSERT INTO "public"."collection_index_item" VALUES ('112226', '7168', '21923');
INSERT INTO "public"."collection_index_item" VALUES ('112229', '7169', '21925');
INSERT INTO "public"."collection_index_item" VALUES ('112230', '7170', '21927');
INSERT INTO "public"."collection_index_item" VALUES ('112231', '7170', '21926');
INSERT INTO "public"."collection_index_item" VALUES ('112232', '7171', '21930');
INSERT INTO "public"."collection_index_item" VALUES ('112233', '7172', '21930');
INSERT INTO "public"."collection_index_item" VALUES ('112234', '7172', '21932');
INSERT INTO "public"."collection_index_item" VALUES ('112235', '7173', '21929');
INSERT INTO "public"."collection_index_item" VALUES ('112245', '7174', '21934');
INSERT INTO "public"."collection_index_item" VALUES ('112246', '7175', '21933');
INSERT INTO "public"."collection_index_item" VALUES ('112249', '7176', '21939');
INSERT INTO "public"."collection_index_item" VALUES ('112250', '7177', '21938');
INSERT INTO "public"."collection_index_item" VALUES ('112256', '7178', '21947');
INSERT INTO "public"."collection_index_item" VALUES ('112257', '7179', '21946');
INSERT INTO "public"."collection_index_item" VALUES ('112258', '7180', '21949');
INSERT INTO "public"."collection_index_item" VALUES ('112259', '7180', '21954');
INSERT INTO "public"."collection_index_item" VALUES ('112260', '7181', '21951');
INSERT INTO "public"."collection_index_item" VALUES ('112261', '7182', '21954');
INSERT INTO "public"."collection_index_item" VALUES ('112262', '7182', '21953');
INSERT INTO "public"."collection_index_item" VALUES ('112263', '7182', '21956');
INSERT INTO "public"."collection_index_item" VALUES ('112264', '7182', '21955');
INSERT INTO "public"."collection_index_item" VALUES ('112265', '7182', '21952');
INSERT INTO "public"."collection_index_item" VALUES ('112266', '7182', '21949');
INSERT INTO "public"."collection_index_item" VALUES ('112267', '7183', '21958');
INSERT INTO "public"."collection_index_item" VALUES ('112268', '7183', '21960');
INSERT INTO "public"."collection_index_item" VALUES ('112269', '7184', '21957');
INSERT INTO "public"."collection_index_item" VALUES ('112284', '7185', '21974');
INSERT INTO "public"."collection_index_item" VALUES ('112285', '7186', '21967');
INSERT INTO "public"."collection_index_item" VALUES ('112286', '7186', '21970');
INSERT INTO "public"."collection_index_item" VALUES ('112287', '7187', '21977');
INSERT INTO "public"."collection_index_item" VALUES ('112288', '7188', '21975');
INSERT INTO "public"."collection_index_item" VALUES ('112355', '7189', '21984');
INSERT INTO "public"."collection_index_item" VALUES ('112461', '7190', '21964');
INSERT INTO "public"."collection_index_item" VALUES ('112462', '7190', '21966');
INSERT INTO "public"."collection_index_item" VALUES ('112463', '7191', '21965');
INSERT INTO "public"."collection_index_item" VALUES ('112464', '7191', '21964');
INSERT INTO "public"."collection_index_item" VALUES ('112465', '7192', '21963');
INSERT INTO "public"."collection_index_item" VALUES ('112502', '7193', '21990');
INSERT INTO "public"."collection_index_item" VALUES ('112503', '7193', '21989');
INSERT INTO "public"."collection_index_item" VALUES ('112504', '7194', '21988');
INSERT INTO "public"."collection_index_item" VALUES ('112509', '7195', '21978');
INSERT INTO "public"."collection_index_item" VALUES ('112510', '7195', '21981');
INSERT INTO "public"."collection_index_item" VALUES ('112511', '7196', '21982');

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
INSERT INTO "public"."database" VALUES ('1628', 'example_forum', '0');
INSERT INTO "public"."database" VALUES ('1904', 'dataman_router', '3');
INSERT INTO "public"."database" VALUES ('1905', 'dataman_storage', '3');

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

-- ----------------------------
-- Table structure for field_type_constraint
-- ----------------------------
DROP TABLE IF EXISTS "public"."field_type_constraint";
CREATE TABLE "public"."field_type_constraint" (
"f" int4 DEFAULT nextval('field_type_constraint__id_seq'::regclass) NOT NULL,
"field_type_id" int4 NOT NULL,
"constraint" varchar(255) COLLATE "default" NOT NULL,
"args" jsonb,
"validation_error" varchar(255) COLLATE "default"
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of field_type_constraint
-- ----------------------------

-- ----------------------------
-- Table structure for shard_instance
-- ----------------------------
DROP TABLE IF EXISTS "public"."shard_instance";
CREATE TABLE "public"."shard_instance" (
"_id" int4 DEFAULT nextval('shard_instance__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"database_id" int4 NOT NULL,
"count" int4,
"instance" int4,
"database_shard" bool NOT NULL,
"collection_shard" bool NOT NULL,
"provision_state" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Records of shard_instance
-- ----------------------------
INSERT INTO "public"."shard_instance" VALUES ('1721', 'dbshard_example_forum_2', '1628', '2', '1', 't', 'f', '0');
INSERT INTO "public"."shard_instance" VALUES ('2158', 'public', '1904', '1', '1', 't', 'f', '3');
INSERT INTO "public"."shard_instance" VALUES ('2159', 'public', '1905', '1', '1', 't', 'f', '3');

-- ----------------------------
-- Alter Sequences Owned By 
-- ----------------------------
ALTER SEQUENCE "public"."collection__id_seq" OWNED BY "collection"."_id";
ALTER SEQUENCE "public"."collection_field__id_seq" OWNED BY "collection_field"."_id";
ALTER SEQUENCE "public"."collection_field_relation__id_seq" OWNED BY "collection_field_relation"."_id";
ALTER SEQUENCE "public"."collection_index__id_seq" OWNED BY "collection_index"."_id";
ALTER SEQUENCE "public"."collection_index_item__id_seq" OWNED BY "collection_index_item"."_id";
ALTER SEQUENCE "public"."database__id_seq" OWNED BY "database"."_id";
ALTER SEQUENCE "public"."shard_instance__id_seq" OWNED BY "shard_instance"."_id";

-- ----------------------------
-- Indexes structure for table collection
-- ----------------------------
CREATE UNIQUE INDEX "collection_name_shard_instance_id_idx" ON "public"."collection" USING btree ("name", "shard_instance_id");

-- ----------------------------
-- Primary Key structure for table collection
-- ----------------------------
ALTER TABLE "public"."collection" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_field
-- ----------------------------
CREATE UNIQUE INDEX "index_collection_field_collection_field_name" ON "public"."collection_field" USING btree ("collection_id", "name");

-- ----------------------------
-- Primary Key structure for table collection_field
-- ----------------------------
ALTER TABLE "public"."collection_field" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Primary Key structure for table collection_field_relation
-- ----------------------------
ALTER TABLE "public"."collection_field_relation" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_index
-- ----------------------------
CREATE UNIQUE INDEX "collection_index_name" ON "public"."collection_index" USING btree ("name", "collection_id");
CREATE UNIQUE INDEX "collection_index_collection_id_primary_idx" ON "public"."collection_index" USING btree ("collection_id", "primary");

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
-- Indexes structure for table database
-- ----------------------------
CREATE UNIQUE INDEX "database_name_idx" ON "public"."database" USING btree ("name");

-- ----------------------------
-- Primary Key structure for table database
-- ----------------------------
ALTER TABLE "public"."database" ADD PRIMARY KEY ("_id");

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
ALTER TABLE "public"."field_type_constraint" ADD PRIMARY KEY ("f");

-- ----------------------------
-- Indexes structure for table shard_instance
-- ----------------------------
CREATE UNIQUE INDEX "shard_instance_database_id_count_instance_database_shard_co_idx" ON "public"."shard_instance" USING btree ("database_id", "count", "instance", "database_shard", "collection_shard", "name");
CREATE UNIQUE INDEX "shard_instance_name_database_id_idx" ON "public"."shard_instance" USING btree ("name", "database_id");

-- ----------------------------
-- Primary Key structure for table shard_instance
-- ----------------------------
ALTER TABLE "public"."shard_instance" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Foreign Key structure for table "public"."collection"
-- ----------------------------
ALTER TABLE "public"."collection" ADD FOREIGN KEY ("shard_instance_id") REFERENCES "public"."shard_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_field"
-- ----------------------------
ALTER TABLE "public"."collection_field" ADD FOREIGN KEY ("collection_id") REFERENCES "public"."collection" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_field_relation"
-- ----------------------------
ALTER TABLE "public"."collection_field_relation" ADD FOREIGN KEY ("relation_collection_field_id") REFERENCES "public"."collection_field" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."collection_field_relation" ADD FOREIGN KEY ("collection_field_id") REFERENCES "public"."collection_field" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_index"
-- ----------------------------
ALTER TABLE "public"."collection_index" ADD FOREIGN KEY ("collection_id") REFERENCES "public"."collection" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_index_item"
-- ----------------------------
ALTER TABLE "public"."collection_index_item" ADD FOREIGN KEY ("collection_index_id") REFERENCES "public"."collection_index" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."collection_index_item" ADD FOREIGN KEY ("collection_field_id") REFERENCES "public"."collection_field" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."shard_instance"
-- ----------------------------
ALTER TABLE "public"."shard_instance" ADD FOREIGN KEY ("database_id") REFERENCES "public"."database" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
