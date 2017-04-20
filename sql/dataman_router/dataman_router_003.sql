/*
Navicat PGSQL Data Transfer

Source Server         : localhost_5432
Source Server Version : 90505
Source Host           : localhost:5432
Source Database       : dataman_router
Source Schema         : public

Target Server Type    : PGSQL
Target Server Version : 90505
File Encoding         : 65001

Date: 2017-03-12 21:56:24
*/


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
-- Sequence structure for database_tombstone_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database_tombstone_id_seq";
CREATE SEQUENCE "public"."database_tombstone_id_seq"
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
-- Sequence structure for schema_item_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."schema_item_id_seq";
CREATE SEQUENCE "public"."schema_item_id_seq"
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
-- Sequence structure for table_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."table_id_seq";
CREATE SEQUENCE "public"."table_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for table_index_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."table_index_id_seq";
CREATE SEQUENCE "public"."table_index_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for tombstone_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."tombstone_id_seq";
CREATE SEQUENCE "public"."tombstone_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Table structure for database
-- ----------------------------
DROP TABLE IF EXISTS "public"."database";
CREATE TABLE "public"."database" (
"id" int4 DEFAULT nextval('database_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"datastore_id" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for database_tombstone
-- ----------------------------
DROP TABLE IF EXISTS "public"."database_tombstone";
CREATE TABLE "public"."database_tombstone" (
"id" int4 DEFAULT nextval('database_tombstone_id_seq'::regclass) NOT NULL,
"database_id" int4,
"tombstone_id" int4,
"ttl" int4
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
"replica_config_json" text COLLATE "default",
"shard_config_json" text COLLATE "default"
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
"info" varchar(255) COLLATE "default"
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for schema_item
-- ----------------------------
DROP TABLE IF EXISTS "public"."schema_item";
CREATE TABLE "public"."schema_item" (
"id" int4 DEFAULT nextval('schema_item_id_seq'::regclass) NOT NULL,
"schema_id" int4,
"version" int4 NOT NULL,
"schema_json" text COLLATE "default",
"backwards_compatible" int2
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
"config_json" text COLLATE "default"
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
"config_json_schema_item_id" int4
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for table
-- ----------------------------
DROP TABLE IF EXISTS "public"."table";
CREATE TABLE "public"."table" (
"id" int4 DEFAULT nextval('table_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"document_schema_item_id" int4,
"database_id" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for table_index
-- ----------------------------
DROP TABLE IF EXISTS "public"."table_index";
CREATE TABLE "public"."table_index" (
"id" int4 DEFAULT nextval('table_index_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"table_id" int4,
"data_json" text COLLATE "default"
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for tombstone
-- ----------------------------
DROP TABLE IF EXISTS "public"."tombstone";
CREATE TABLE "public"."tombstone" (
"id" int4 DEFAULT nextval('tombstone_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"info" varchar(255) COLLATE "default"
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Alter Sequences Owned By 
-- ----------------------------
ALTER SEQUENCE "public"."database_id_seq" OWNED BY "database"."id";
ALTER SEQUENCE "public"."database_tombstone_id_seq" OWNED BY "database_tombstone"."id";
ALTER SEQUENCE "public"."datastore_id_seq" OWNED BY "datastore"."id";
ALTER SEQUENCE "public"."datastore_shard_id_seq" OWNED BY "datastore_shard"."id";
ALTER SEQUENCE "public"."datastore_shard_item_id_seq" OWNED BY "datastore_shard_item"."id";
ALTER SEQUENCE "public"."schema_id_seq" OWNED BY "schema"."id";
ALTER SEQUENCE "public"."schema_item_id_seq" OWNED BY "schema_item"."id";
ALTER SEQUENCE "public"."storage_node_id_seq" OWNED BY "storage_node"."id";
ALTER SEQUENCE "public"."storage_node_state_id_seq" OWNED BY "storage_node_state"."id";
ALTER SEQUENCE "public"."storage_node_type_id_seq" OWNED BY "storage_node_type"."id";
ALTER SEQUENCE "public"."table_id_seq" OWNED BY "table"."id";
ALTER SEQUENCE "public"."table_index_id_seq" OWNED BY "table_index"."id";
ALTER SEQUENCE "public"."tombstone_id_seq" OWNED BY "tombstone"."id";

-- ----------------------------
-- Indexes structure for table database
-- ----------------------------
CREATE INDEX "datastore_id_idx" ON "public"."database" USING btree ("datastore_id");

-- ----------------------------
-- Primary Key structure for table database
-- ----------------------------
ALTER TABLE "public"."database" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table database_tombstone
-- ----------------------------
ALTER TABLE "public"."database_tombstone" ADD PRIMARY KEY ("id");

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
-- Primary Key structure for table schema
-- ----------------------------
ALTER TABLE "public"."schema" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table schema_item
-- ----------------------------
ALTER TABLE "public"."schema_item" ADD PRIMARY KEY ("id");

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
-- Indexes structure for table table
-- ----------------------------
CREATE UNIQUE INDEX "database_table" ON "public"."table" USING btree ("database_id", "name");

-- ----------------------------
-- Primary Key structure for table table
-- ----------------------------
ALTER TABLE "public"."table" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table table_index
-- ----------------------------
CREATE UNIQUE INDEX "table_index_uniq" ON "public"."table_index" USING btree ("table_id", "name");

-- ----------------------------
-- Primary Key structure for table table_index
-- ----------------------------
ALTER TABLE "public"."table_index" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table tombstone
-- ----------------------------
ALTER TABLE "public"."tombstone" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Foreign Key structure for table "public"."database"
-- ----------------------------
ALTER TABLE "public"."database" ADD FOREIGN KEY ("datastore_id") REFERENCES "public"."datastore" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."database_tombstone"
-- ----------------------------
ALTER TABLE "public"."database_tombstone" ADD FOREIGN KEY ("database_id") REFERENCES "public"."database" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."database_tombstone" ADD FOREIGN KEY ("tombstone_id") REFERENCES "public"."tombstone" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

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
-- Foreign Key structure for table "public"."schema_item"
-- ----------------------------
ALTER TABLE "public"."schema_item" ADD FOREIGN KEY ("schema_id") REFERENCES "public"."schema" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."storage_node"
-- ----------------------------
ALTER TABLE "public"."storage_node" ADD FOREIGN KEY ("storage_node_state_id") REFERENCES "public"."storage_node_state" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."storage_node" ADD FOREIGN KEY ("storage_node_type_id") REFERENCES "public"."storage_node_type" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."storage_node_type"
-- ----------------------------
ALTER TABLE "public"."storage_node_type" ADD FOREIGN KEY ("config_json_schema_item_id") REFERENCES "public"."schema_item" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."table"
-- ----------------------------
ALTER TABLE "public"."table" ADD FOREIGN KEY ("database_id") REFERENCES "public"."database" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."table" ADD FOREIGN KEY ("document_schema_item_id") REFERENCES "public"."schema_item" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."table_index"
-- ----------------------------
ALTER TABLE "public"."table_index" ADD FOREIGN KEY ("table_id") REFERENCES "public"."table" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
