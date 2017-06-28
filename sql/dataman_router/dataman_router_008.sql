/*
Navicat PGSQL Data Transfer

Source Server         : local
Source Server Version : 90603
Source Host           : localhost:5432
Source Database       : dataman_router
Source Schema         : public

Target Server Type    : PGSQL
Target Server Version : 90603
File Encoding         : 65001

Date: 2017-06-28 11:33:14
*/


-- ----------------------------
-- Sequence structure for collection__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection__id_seq";
CREATE SEQUENCE "public"."collection__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 4185
 CACHE 1;
SELECT setval('"public"."collection__id_seq"', 4185, true);

-- ----------------------------
-- Sequence structure for collection_field__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_field__id_seq";
CREATE SEQUENCE "public"."collection_field__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 17435
 CACHE 1;
SELECT setval('"public"."collection_field__id_seq"', 17435, true);

-- ----------------------------
-- Sequence structure for collection_field_relation__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_field_relation__id_seq";
CREATE SEQUENCE "public"."collection_field_relation__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1250
 CACHE 1;
SELECT setval('"public"."collection_field_relation__id_seq"', 1250, true);

-- ----------------------------
-- Sequence structure for collection_index__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_index__id_seq";
CREATE SEQUENCE "public"."collection_index__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 5444
 CACHE 1;
SELECT setval('"public"."collection_index__id_seq"', 5444, true);

-- ----------------------------
-- Sequence structure for collection_index_item__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_index_item__id_seq";
CREATE SEQUENCE "public"."collection_index_item__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 61152
 CACHE 1;
SELECT setval('"public"."collection_index_item__id_seq"', 61152, true);

-- ----------------------------
-- Sequence structure for collection_keyspace__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_keyspace__id_seq";
CREATE SEQUENCE "public"."collection_keyspace__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for collection_keyspace_item__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_keyspace_item__id_seq";
CREATE SEQUENCE "public"."collection_keyspace_item__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for collection_keyspace_partition__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_keyspace_partition__id_seq";
CREATE SEQUENCE "public"."collection_keyspace_partition__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for collection_keyspace_partition_datastore_vshard__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_keyspace_partition_datastore_vshard__id_seq";
CREATE SEQUENCE "public"."collection_keyspace_partition_datastore_vshard__id_seq"
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
 START 1449
 CACHE 1;
SELECT setval('"public"."database__id_seq"', 1449, true);

-- ----------------------------
-- Sequence structure for database_datastore__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database_datastore__id_seq";
CREATE SEQUENCE "public"."database_datastore__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1382
 CACHE 1;
SELECT setval('"public"."database_datastore__id_seq"', 1382, true);

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
 START 3
 CACHE 1;
SELECT setval('"public"."datasource__id_seq"', 3, true);

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
 START 117
 CACHE 1;
SELECT setval('"public"."datasource_instance__id_seq"', 117, true);

-- ----------------------------
-- Sequence structure for datasource_instance_shard_instance__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datasource_instance_shard_instance__id_seq";
CREATE SEQUENCE "public"."datasource_instance_shard_instance__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 3271
 CACHE 1;
SELECT setval('"public"."datasource_instance_shard_instance__id_seq"', 3271, true);

-- ----------------------------
-- Sequence structure for datastore__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datastore__id_seq";
CREATE SEQUENCE "public"."datastore__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 55
 CACHE 1;
SELECT setval('"public"."datastore__id_seq"', 55, true);

-- ----------------------------
-- Sequence structure for datastore_shard__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datastore_shard__id_seq";
CREATE SEQUENCE "public"."datastore_shard__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 69
 CACHE 1;
SELECT setval('"public"."datastore_shard__id_seq"', 69, true);

-- ----------------------------
-- Sequence structure for datastore_shard_replica__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datastore_shard_replica__id_seq";
CREATE SEQUENCE "public"."datastore_shard_replica__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 68
 CACHE 1;
SELECT setval('"public"."datastore_shard_replica__id_seq"', 68, true);

-- ----------------------------
-- Sequence structure for datastore_vshard__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datastore_vshard__id_seq";
CREATE SEQUENCE "public"."datastore_vshard__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;
SELECT setval('"public"."datastore_vshard__id_seq"', 1, true);

-- ----------------------------
-- Sequence structure for datastore_vshard_instance__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datastore_vshard_instance__id_seq";
CREATE SEQUENCE "public"."datastore_vshard_instance__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for datastore_vshard_instance_datastore_shard__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datastore_vshard_instance_datastore_shard__id_seq";
CREATE SEQUENCE "public"."datastore_vshard_instance_datastore_shard__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 2
 CACHE 1;
SELECT setval('"public"."datastore_vshard_instance_datastore_shard__id_seq"', 2, true);

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
-- Sequence structure for storage_node_type__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."storage_node_type__id_seq";
CREATE SEQUENCE "public"."storage_node_type__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 138
 CACHE 1;
SELECT setval('"public"."storage_node_type__id_seq"', 138, true);

-- ----------------------------
-- Table structure for collection
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection";
CREATE TABLE "public"."collection" (
"_id" int4 DEFAULT nextval('collection__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"database_id" int4,
"provision_state" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for collection_field
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_field";
CREATE TABLE "public"."collection_field" (
"_id" int4 DEFAULT nextval('collection_field__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"collection_id" int4,
"field_type" varchar(255) COLLATE "default",
"not_null" bool NOT NULL,
"parent_collection_field_id" int4,
"provision_state" int4 NOT NULL,
"default" varchar(255) COLLATE "default"
)
WITH (OIDS=FALSE)

;

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
-- Table structure for collection_keyspace
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_keyspace";
CREATE TABLE "public"."collection_keyspace" (
"_id" int4 DEFAULT nextval('collection_keyspace__id_seq'::regclass) NOT NULL,
"collection_id" int4 NOT NULL,
"hash_method" varchar(255) COLLATE "default" NOT NULL,
"write" bool NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for collection_keyspace_item
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_keyspace_item";
CREATE TABLE "public"."collection_keyspace_item" (
"_id" int4 DEFAULT nextval('collection_keyspace_item__id_seq'::regclass) NOT NULL,
"collection_keyspace_id" int4 NOT NULL,
"collection_field_id" int4 NOT NULL,
"order" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for collection_keyspace_partition
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_keyspace_partition";
CREATE TABLE "public"."collection_keyspace_partition" (
"_id" int4 DEFAULT nextval('collection_keyspace_partition__id_seq'::regclass) NOT NULL,
"collection_keyspace_id" int4 NOT NULL,
"start_id" int4 NOT NULL,
"end_id" int4,
"shard_method" varchar(255) COLLATE "default" NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for collection_keyspace_partition_datastore_vshard_instance
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_keyspace_partition_datastore_vshard_instance";
CREATE TABLE "public"."collection_keyspace_partition_datastore_vshard_instance" (
"_id" int4 DEFAULT nextval('collection_keyspace_partition_datastore_vshard__id_seq'::regclass) NOT NULL,
"collection_keyspace_partition_id" int4 NOT NULL,
"datastore_vshard_instance_id" int4 NOT NULL,
"order" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

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
"provision_state" int4 NOT NULL,
"datastore_vshard_id" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

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
-- Table structure for datasource_instance_shard_instance
-- ----------------------------
DROP TABLE IF EXISTS "public"."datasource_instance_shard_instance";
CREATE TABLE "public"."datasource_instance_shard_instance" (
"_id" int4 DEFAULT nextval('datasource_instance_shard_instance__id_seq'::regclass) NOT NULL,
"datasource_instance_id" int4,
"datastore_vshard_instance_id" int4,
"name" varchar(255) COLLATE "default",
"provision_state" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

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
-- Table structure for datastore_vshard
-- ----------------------------
DROP TABLE IF EXISTS "public"."datastore_vshard";
CREATE TABLE "public"."datastore_vshard" (
"_id" int4 DEFAULT nextval('datastore_vshard__id_seq'::regclass) NOT NULL,
"datastore_id" int4 NOT NULL,
"shard_count" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for datastore_vshard_instance
-- ----------------------------
DROP TABLE IF EXISTS "public"."datastore_vshard_instance";
CREATE TABLE "public"."datastore_vshard_instance" (
"_id" int4 DEFAULT nextval('datastore_vshard_instance__id_seq'::regclass) NOT NULL,
"datastore_vshard_id" int4 NOT NULL,
"shard_instance" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for datastore_vshard_instance_datastore_shard
-- ----------------------------
DROP TABLE IF EXISTS "public"."datastore_vshard_instance_datastore_shard";
CREATE TABLE "public"."datastore_vshard_instance_datastore_shard" (
"_id" int4 DEFAULT nextval('datastore_vshard_instance_datastore_shard__id_seq'::regclass) NOT NULL,
"datastore_vshard_instance_id" int4 NOT NULL,
"datastore_shard_id" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

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
-- Alter Sequences Owned By 
-- ----------------------------
ALTER SEQUENCE "public"."collection__id_seq" OWNED BY "collection"."_id";
ALTER SEQUENCE "public"."collection_field__id_seq" OWNED BY "collection_field"."_id";
ALTER SEQUENCE "public"."collection_field_relation__id_seq" OWNED BY "collection_field_relation"."_id";
ALTER SEQUENCE "public"."collection_index__id_seq" OWNED BY "collection_index"."_id";
ALTER SEQUENCE "public"."collection_index_item__id_seq" OWNED BY "collection_index_item"."_id";
ALTER SEQUENCE "public"."collection_keyspace__id_seq" OWNED BY "collection_keyspace"."_id";
ALTER SEQUENCE "public"."collection_keyspace_item__id_seq" OWNED BY "collection_keyspace_item"."_id";
ALTER SEQUENCE "public"."collection_keyspace_partition__id_seq" OWNED BY "collection_keyspace_partition"."_id";
ALTER SEQUENCE "public"."collection_keyspace_partition_datastore_vshard__id_seq" OWNED BY "collection_keyspace_partition_datastore_vshard_instance"."_id";
ALTER SEQUENCE "public"."database__id_seq" OWNED BY "database"."_id";
ALTER SEQUENCE "public"."database_datastore__id_seq" OWNED BY "database_datastore"."_id";
ALTER SEQUENCE "public"."datasource__id_seq" OWNED BY "datasource"."_id";
ALTER SEQUENCE "public"."datasource_instance__id_seq" OWNED BY "datasource_instance"."_id";
ALTER SEQUENCE "public"."datasource_instance_shard_instance__id_seq" OWNED BY "datasource_instance_shard_instance"."_id";
ALTER SEQUENCE "public"."datastore__id_seq" OWNED BY "datastore"."_id";
ALTER SEQUENCE "public"."datastore_shard__id_seq" OWNED BY "datastore_shard"."_id";
ALTER SEQUENCE "public"."datastore_shard_replica__id_seq" OWNED BY "datastore_shard_replica"."_id";
ALTER SEQUENCE "public"."datastore_vshard__id_seq" OWNED BY "datastore_vshard"."_id";
ALTER SEQUENCE "public"."datastore_vshard_instance__id_seq" OWNED BY "datastore_vshard_instance"."_id";
ALTER SEQUENCE "public"."datastore_vshard_instance_datastore_shard__id_seq" OWNED BY "datastore_vshard_instance_datastore_shard"."_id";
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
-- Primary Key structure for table collection_keyspace
-- ----------------------------
ALTER TABLE "public"."collection_keyspace" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_keyspace_item
-- ----------------------------
CREATE UNIQUE INDEX "collection_keyspace_item_collection_keyspace_id_order_idx" ON "public"."collection_keyspace_item" USING btree ("collection_keyspace_id", "order");

-- ----------------------------
-- Primary Key structure for table collection_keyspace_item
-- ----------------------------
ALTER TABLE "public"."collection_keyspace_item" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Primary Key structure for table collection_keyspace_partition
-- ----------------------------
ALTER TABLE "public"."collection_keyspace_partition" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Primary Key structure for table collection_keyspace_partition_datastore_vshard_instance
-- ----------------------------
ALTER TABLE "public"."collection_keyspace_partition_datastore_vshard_instance" ADD PRIMARY KEY ("_id");

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
-- Primary Key structure for table datastore_vshard
-- ----------------------------
ALTER TABLE "public"."datastore_vshard" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table datastore_vshard_instance
-- ----------------------------
CREATE UNIQUE INDEX "datastore_vshard_instance_datastore_vshard_id_shard_instanc_idx" ON "public"."datastore_vshard_instance" USING btree ("datastore_vshard_id", "shard_instance");

-- ----------------------------
-- Primary Key structure for table datastore_vshard_instance
-- ----------------------------
ALTER TABLE "public"."datastore_vshard_instance" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table datastore_vshard_instance_datastore_shard
-- ----------------------------
CREATE UNIQUE INDEX "datastore_vshard_instance_data_datastore_vshard_instance_id_idx" ON "public"."datastore_vshard_instance_datastore_shard" USING btree ("datastore_vshard_instance_id");

-- ----------------------------
-- Primary Key structure for table datastore_vshard_instance_datastore_shard
-- ----------------------------
ALTER TABLE "public"."datastore_vshard_instance_datastore_shard" ADD PRIMARY KEY ("_id");

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
ALTER TABLE "public"."collection" ADD FOREIGN KEY ("database_id") REFERENCES "public"."database" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_field"
-- ----------------------------
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
ALTER TABLE "public"."collection_index_item" ADD FOREIGN KEY ("collection_index_id") REFERENCES "public"."collection_index" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."collection_index_item" ADD FOREIGN KEY ("collection_field_id") REFERENCES "public"."collection_field" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_keyspace"
-- ----------------------------
ALTER TABLE "public"."collection_keyspace" ADD FOREIGN KEY ("collection_id") REFERENCES "public"."collection" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_keyspace_item"
-- ----------------------------
ALTER TABLE "public"."collection_keyspace_item" ADD FOREIGN KEY ("collection_keyspace_id") REFERENCES "public"."collection_keyspace" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."collection_keyspace_item" ADD FOREIGN KEY ("collection_field_id") REFERENCES "public"."collection_field" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_keyspace_partition"
-- ----------------------------
ALTER TABLE "public"."collection_keyspace_partition" ADD FOREIGN KEY ("collection_keyspace_id") REFERENCES "public"."collection_keyspace" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_keyspace_partition_datastore_vshard_instance"
-- ----------------------------
ALTER TABLE "public"."collection_keyspace_partition_datastore_vshard_instance" ADD FOREIGN KEY ("collection_keyspace_partition_id") REFERENCES "public"."collection_keyspace_partition" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."collection_keyspace_partition_datastore_vshard_instance" ADD FOREIGN KEY ("datastore_vshard_instance_id") REFERENCES "public"."datastore_vshard_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."database_datastore"
-- ----------------------------
ALTER TABLE "public"."database_datastore" ADD FOREIGN KEY ("database_id") REFERENCES "public"."database" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."database_datastore" ADD FOREIGN KEY ("datastore_vshard_id") REFERENCES "public"."datastore_shard" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."database_datastore" ADD FOREIGN KEY ("datastore_id") REFERENCES "public"."datastore" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datasource_instance"
-- ----------------------------
ALTER TABLE "public"."datasource_instance" ADD FOREIGN KEY ("storage_node_id") REFERENCES "public"."storage_node" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."datasource_instance" ADD FOREIGN KEY ("datasource_id") REFERENCES "public"."datasource" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datasource_instance_shard_instance"
-- ----------------------------
ALTER TABLE "public"."datasource_instance_shard_instance" ADD FOREIGN KEY ("datasource_instance_id") REFERENCES "public"."datasource_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."datasource_instance_shard_instance" ADD FOREIGN KEY ("datastore_vshard_instance_id") REFERENCES "public"."datastore_vshard_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datastore_shard"
-- ----------------------------
ALTER TABLE "public"."datastore_shard" ADD FOREIGN KEY ("datastore_id") REFERENCES "public"."datastore" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datastore_shard_replica"
-- ----------------------------
ALTER TABLE "public"."datastore_shard_replica" ADD FOREIGN KEY ("datasource_instance_id") REFERENCES "public"."datasource_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."datastore_shard_replica" ADD FOREIGN KEY ("datastore_shard_id") REFERENCES "public"."datastore_shard" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datastore_vshard"
-- ----------------------------
ALTER TABLE "public"."datastore_vshard" ADD FOREIGN KEY ("datastore_id") REFERENCES "public"."datastore" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datastore_vshard_instance"
-- ----------------------------
ALTER TABLE "public"."datastore_vshard_instance" ADD FOREIGN KEY ("datastore_vshard_id") REFERENCES "public"."datastore_vshard" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datastore_vshard_instance_datastore_shard"
-- ----------------------------
ALTER TABLE "public"."datastore_vshard_instance_datastore_shard" ADD FOREIGN KEY ("datastore_vshard_instance_id") REFERENCES "public"."datastore_vshard_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."datastore_vshard_instance_datastore_shard" ADD FOREIGN KEY ("datastore_shard_id") REFERENCES "public"."datastore_shard" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."field_type_constraint"
-- ----------------------------
ALTER TABLE "public"."field_type_constraint" ADD FOREIGN KEY ("field_type_id") REFERENCES "public"."field_type" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
