/*
Navicat PGSQL Data Transfer

Source Server         : local
Source Server Version : 90506
Source Host           : localhost:5432
Source Database       : dataman_router
Source Schema         : public

Target Server Type    : PGSQL
Target Server Version : 90506
File Encoding         : 65001

Date: 2017-04-12 10:20:20
*/


-- ----------------------------
-- Sequence structure for collection_field_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_field_id_seq";
CREATE SEQUENCE "public"."collection_field_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for collection_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_id_seq";
CREATE SEQUENCE "public"."collection_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for collection_index_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_index_id_seq";
CREATE SEQUENCE "public"."collection_index_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for database_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database_id_seq";
CREATE SEQUENCE "public"."database_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for datastore_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datastore_id_seq";
CREATE SEQUENCE "public"."datastore_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for datastore_shard_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datastore_shard_id_seq";
CREATE SEQUENCE "public"."datastore_shard_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;
SELECT setval('"public"."datastore_shard_id_seq"', 1, true);

-- ----------------------------
-- Sequence structure for datastore_shard_item_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datastore_shard_item_id_seq";
CREATE SEQUENCE "public"."datastore_shard_item_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for schema_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."schema_id_seq";
CREATE SEQUENCE "public"."schema_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for storage_node_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."storage_node_id_seq";
CREATE SEQUENCE "public"."storage_node_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for storage_node_state_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."storage_node_state_id_seq";
CREATE SEQUENCE "public"."storage_node_state_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for storage_node_type_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."storage_node_type_id_seq";
CREATE SEQUENCE "public"."storage_node_type_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Table structure for collection
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection";
CREATE TABLE "public"."collection" (
"id" int4 DEFAULT nextval('collection_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"database_id" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for collection_field
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_field";
CREATE TABLE "public"."collection_field" (
"id" int4 DEFAULT nextval('collection_field_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"collection_id" int4 NOT NULL,
"field_type" varchar(255) COLLATE "default" NOT NULL,
"field_type_args" varchar(255) COLLATE "default",
"order" int4,
"schema_id" int4,
"not_null" bool
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for collection_index
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_index";
CREATE TABLE "public"."collection_index" (
"id" int4 DEFAULT nextval('collection_index_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"collection_id" int4 NOT NULL,
"data_json" jsonb NOT NULL,
"unique" bool
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for database
-- ----------------------------
DROP TABLE IF EXISTS "public"."database";
CREATE TABLE "public"."database" (
"id" int4 DEFAULT nextval('database_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"primary_datastore_id" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for datastore
-- ----------------------------
DROP TABLE IF EXISTS "public"."datastore";
CREATE TABLE "public"."datastore" (
"id" int4 DEFAULT nextval('datastore_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"replica_config_json" jsonb,
"shard_config_json" jsonb
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for datastore_shard
-- ----------------------------
DROP TABLE IF EXISTS "public"."datastore_shard";
CREATE TABLE "public"."datastore_shard" (
"id" int4 DEFAULT nextval('datastore_shard_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"datastore_id" int4
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for datastore_shard_item
-- ----------------------------
DROP TABLE IF EXISTS "public"."datastore_shard_item";
CREATE TABLE "public"."datastore_shard_item" (
"id" int4 DEFAULT nextval('datastore_shard_item_id_seq'::regclass) NOT NULL,
"datastore_shard_id" int4,
"storage_node_id" int4
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for schema
-- ----------------------------
DROP TABLE IF EXISTS "public"."schema";
CREATE TABLE "public"."schema" (
"id" int4 DEFAULT nextval('schema_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"version" int4 NOT NULL,
"data_json" jsonb NOT NULL,
"backwards_compatible" bool
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for storage_node
-- ----------------------------
DROP TABLE IF EXISTS "public"."storage_node";
CREATE TABLE "public"."storage_node" (
"id" int4 DEFAULT nextval('storage_node_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"ip" varchar(255) COLLATE "default",
"port" int4,
"storage_node_type_id" int4,
"storage_node_state_id" int4,
"config_json" jsonb
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for storage_node_state
-- ----------------------------
DROP TABLE IF EXISTS "public"."storage_node_state";
CREATE TABLE "public"."storage_node_state" (
"id" int4 DEFAULT nextval('storage_node_state_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"info" varchar(255) COLLATE "default"
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for storage_node_type
-- ----------------------------
DROP TABLE IF EXISTS "public"."storage_node_type";
CREATE TABLE "public"."storage_node_type" (
"id" int4 DEFAULT nextval('storage_node_type_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"config_json_schema_id" int4
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Alter Sequences Owned By 
-- ----------------------------
ALTER SEQUENCE "public"."collection_field_id_seq" OWNED BY "collection_field"."id";
ALTER SEQUENCE "public"."collection_id_seq" OWNED BY "collection"."id";
ALTER SEQUENCE "public"."collection_index_id_seq" OWNED BY "collection_index"."id";
ALTER SEQUENCE "public"."database_id_seq" OWNED BY "database"."id";
ALTER SEQUENCE "public"."datastore_id_seq" OWNED BY "datastore"."id";
ALTER SEQUENCE "public"."datastore_shard_id_seq" OWNED BY "datastore_shard"."id";
ALTER SEQUENCE "public"."datastore_shard_item_id_seq" OWNED BY "datastore_shard_item"."id";
ALTER SEQUENCE "public"."schema_id_seq" OWNED BY "schema"."id";
ALTER SEQUENCE "public"."storage_node_id_seq" OWNED BY "storage_node"."id";
ALTER SEQUENCE "public"."storage_node_state_id_seq" OWNED BY "storage_node_state"."id";
ALTER SEQUENCE "public"."storage_node_type_id_seq" OWNED BY "storage_node_type"."id";

-- ----------------------------
-- Indexes structure for table collection
-- ----------------------------
CREATE UNIQUE INDEX "database_collection" ON "public"."collection" USING btree ("database_id", "name");

-- ----------------------------
-- Primary Key structure for table collection
-- ----------------------------
ALTER TABLE "public"."collection" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table collection_field
-- ----------------------------
CREATE UNIQUE INDEX "collection_field_name" ON "public"."collection_field" USING btree ("collection_id", "name");

-- ----------------------------
-- Primary Key structure for table collection_field
-- ----------------------------
ALTER TABLE "public"."collection_field" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table collection_index
-- ----------------------------
CREATE UNIQUE INDEX "collection_index_name" ON "public"."collection_index" USING btree ("collection_id", "name");

-- ----------------------------
-- Primary Key structure for table collection_index
-- ----------------------------
ALTER TABLE "public"."collection_index" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table database
-- ----------------------------
CREATE INDEX "datastore_id_idx" ON "public"."database" USING btree ("primary_datastore_id");

-- ----------------------------
-- Primary Key structure for table database
-- ----------------------------
ALTER TABLE "public"."database" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table datastore
-- ----------------------------
ALTER TABLE "public"."datastore" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table datastore_shard
-- ----------------------------
ALTER TABLE "public"."datastore_shard" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table datastore_shard_item
-- ----------------------------
ALTER TABLE "public"."datastore_shard_item" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table schema
-- ----------------------------
CREATE UNIQUE INDEX "name_version" ON "public"."schema" USING btree ("name", "version");

-- ----------------------------
-- Primary Key structure for table schema
-- ----------------------------
ALTER TABLE "public"."schema" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table storage_node
-- ----------------------------
ALTER TABLE "public"."storage_node" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table storage_node_state
-- ----------------------------
ALTER TABLE "public"."storage_node_state" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table storage_node_type
-- ----------------------------
ALTER TABLE "public"."storage_node_type" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Foreign Key structure for table "public"."collection_field"
-- ----------------------------
ALTER TABLE "public"."collection_field" ADD FOREIGN KEY ("schema_id") REFERENCES "public"."schema" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."collection_field" ADD FOREIGN KEY ("collection_id") REFERENCES "public"."collection" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_index"
-- ----------------------------
ALTER TABLE "public"."collection_index" ADD FOREIGN KEY ("collection_id") REFERENCES "public"."collection" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."database"
-- ----------------------------
ALTER TABLE "public"."database" ADD FOREIGN KEY ("primary_datastore_id") REFERENCES "public"."datastore" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datastore_shard"
-- ----------------------------
ALTER TABLE "public"."datastore_shard" ADD FOREIGN KEY ("datastore_id") REFERENCES "public"."datastore" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datastore_shard_item"
-- ----------------------------
ALTER TABLE "public"."datastore_shard_item" ADD FOREIGN KEY ("datastore_shard_id") REFERENCES "public"."datastore_shard" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."datastore_shard_item" ADD FOREIGN KEY ("storage_node_id") REFERENCES "public"."storage_node" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."storage_node"
-- ----------------------------
ALTER TABLE "public"."storage_node" ADD FOREIGN KEY ("storage_node_type_id") REFERENCES "public"."storage_node_type" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."storage_node" ADD FOREIGN KEY ("storage_node_state_id") REFERENCES "public"."storage_node_state" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
