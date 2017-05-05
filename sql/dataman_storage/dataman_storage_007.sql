/*
Navicat PGSQL Data Transfer

Source Server         : local
Source Server Version : 90602
Source Host           : localhost:5432
Source Database       : dataman_storage
Source Schema         : public

Target Server Type    : PGSQL
Target Server Version : 90602
File Encoding         : 65001

Date: 2017-05-05 10:23:00
*/


-- ----------------------------
-- Sequence structure for collection__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection__id_seq";
CREATE SEQUENCE "public"."collection__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 423
 CACHE 1;
SELECT setval('"public"."collection__id_seq"', 423, true);

-- ----------------------------
-- Sequence structure for collection_field__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_field__id_seq";
CREATE SEQUENCE "public"."collection_field__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 587
 CACHE 1;
SELECT setval('"public"."collection_field__id_seq"', 587, true);

-- ----------------------------
-- Sequence structure for collection_index__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_index__id_seq";
CREATE SEQUENCE "public"."collection_index__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 332
 CACHE 1;
SELECT setval('"public"."collection_index__id_seq"', 332, true);

-- ----------------------------
-- Sequence structure for database__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database__id_seq";
CREATE SEQUENCE "public"."database__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 127
 CACHE 1;
SELECT setval('"public"."database__id_seq"', 127, true);

-- ----------------------------
-- Sequence structure for schema__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."schema__id_seq";
CREATE SEQUENCE "public"."schema__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 2
 CACHE 1;
SELECT setval('"public"."schema__id_seq"', 2, true);

-- ----------------------------
-- Sequence structure for shard_instance__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."shard_instance__id_seq";
CREATE SEQUENCE "public"."shard_instance__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 112
 CACHE 1;
SELECT setval('"public"."shard_instance__id_seq"', 112, true);

-- ----------------------------
-- Table structure for collection
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection";
CREATE TABLE "public"."collection" (
"_id" int4 DEFAULT nextval('collection__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"shard_instance_id" int4 NOT NULL
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
"not_null" int4,
"field_type_args" jsonb,
"parent_collection_field_id" int4
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
"data_json" text COLLATE "default",
"unique" bool
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for database
-- ----------------------------
DROP TABLE IF EXISTS "public"."database";
CREATE TABLE "public"."database" (
"_id" int4 DEFAULT nextval('database__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default"
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for schema
-- ----------------------------
DROP TABLE IF EXISTS "public"."schema";
CREATE TABLE "public"."schema" (
"_id" int4 DEFAULT nextval('schema__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"version" int4,
"data_json" jsonb,
"backwards_compatible" bool
)
WITH (OIDS=FALSE)

;

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
"collection_shard" bool NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Alter Sequences Owned By 
-- ----------------------------
ALTER SEQUENCE "public"."collection__id_seq" OWNED BY "collection"."_id";
ALTER SEQUENCE "public"."collection_field__id_seq" OWNED BY "collection_field"."_id";
ALTER SEQUENCE "public"."collection_index__id_seq" OWNED BY "collection_index"."_id";
ALTER SEQUENCE "public"."database__id_seq" OWNED BY "database"."_id";
ALTER SEQUENCE "public"."schema__id_seq" OWNED BY "schema"."_id";
ALTER SEQUENCE "public"."shard_instance__id_seq" OWNED BY "shard_instance"."_id";

-- ----------------------------
-- Primary Key structure for table collection
-- ----------------------------
ALTER TABLE "public"."collection" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_field
-- ----------------------------
CREATE INDEX "index_collection_field_collection_field_table" ON "public"."collection_field" USING btree ("collection_id");
CREATE UNIQUE INDEX "index_collection_field_collection_field_name" ON "public"."collection_field" USING btree ("collection_id", "name");

-- ----------------------------
-- Primary Key structure for table collection_field
-- ----------------------------
ALTER TABLE "public"."collection_field" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_index
-- ----------------------------
CREATE UNIQUE INDEX "collection_index_name" ON "public"."collection_index" USING btree ("name", "collection_id");

-- ----------------------------
-- Primary Key structure for table collection_index
-- ----------------------------
ALTER TABLE "public"."collection_index" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table database
-- ----------------------------
CREATE UNIQUE INDEX "database_name_idx" ON "public"."database" USING btree ("name");

-- ----------------------------
-- Primary Key structure for table database
-- ----------------------------
ALTER TABLE "public"."database" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Primary Key structure for table schema
-- ----------------------------
ALTER TABLE "public"."schema" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table shard_instance
-- ----------------------------
CREATE UNIQUE INDEX "shard_instance_database_id_count_instance_database_shard_co_idx" ON "public"."shard_instance" USING btree ("database_id", "count", "instance", "database_shard", "collection_shard");
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
ALTER TABLE "public"."collection_field" ADD FOREIGN KEY ("parent_collection_field_id") REFERENCES "public"."collection_field" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_index"
-- ----------------------------
ALTER TABLE "public"."collection_index" ADD FOREIGN KEY ("collection_id") REFERENCES "public"."collection" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."shard_instance"
-- ----------------------------
ALTER TABLE "public"."shard_instance" ADD FOREIGN KEY ("database_id") REFERENCES "public"."database" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
