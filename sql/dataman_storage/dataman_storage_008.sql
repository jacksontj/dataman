/*
Navicat PGSQL Data Transfer

Source Server         : local postgres
Source Server Version : 90506
Source Host           : localhost:5432
Source Database       : dataman_storage
Source Schema         : public

Target Server Type    : PGSQL
Target Server Version : 90506
File Encoding         : 65001

Date: 2017-06-07 19:13:31
*/


-- ----------------------------
-- Sequence structure for collection__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection__id_seq";
CREATE SEQUENCE "public"."collection__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 773
 CACHE 1;
SELECT setval('"public"."collection__id_seq"', 773, true);

-- ----------------------------
-- Sequence structure for collection_field__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_field__id_seq";
CREATE SEQUENCE "public"."collection_field__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1972
 CACHE 1;
SELECT setval('"public"."collection_field__id_seq"', 1972, true);

-- ----------------------------
-- Sequence structure for collection_field_relation__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_field_relation__id_seq";
CREATE SEQUENCE "public"."collection_field_relation__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 62
 CACHE 1;
SELECT setval('"public"."collection_field_relation__id_seq"', 62, true);

-- ----------------------------
-- Sequence structure for collection_index__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_index__id_seq";
CREATE SEQUENCE "public"."collection_index__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 745
 CACHE 1;
SELECT setval('"public"."collection_index__id_seq"', 745, true);

-- ----------------------------
-- Sequence structure for collection_index_item__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_index_item__id_seq";
CREATE SEQUENCE "public"."collection_index_item__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 628
 CACHE 1;
SELECT setval('"public"."collection_index_item__id_seq"', 628, true);

-- ----------------------------
-- Sequence structure for database__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database__id_seq";
CREATE SEQUENCE "public"."database__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 251
 CACHE 1;
SELECT setval('"public"."database__id_seq"', 251, true);

-- ----------------------------
-- Sequence structure for shard_instance__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."shard_instance__id_seq";
CREATE SEQUENCE "public"."shard_instance__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 257
 CACHE 1;
SELECT setval('"public"."shard_instance__id_seq"', 257, true);

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
-- Table structure for collection_field
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_field";
CREATE TABLE "public"."collection_field" (
"_id" int4 DEFAULT nextval('collection_field__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"collection_id" int4,
"field_type" varchar(255) COLLATE "default",
"not_null" int4,
"parent_collection_field_id" int4,
"provision_state" int4 NOT NULL
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
"cascade_on_delete" bool NOT NULL,
"provision_state" int4 NOT NULL
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
-- Indexes structure for table shard_instance
-- ----------------------------
CREATE UNIQUE INDEX "shard_instance_name_database_id_idx" ON "public"."shard_instance" USING btree ("name", "database_id");
CREATE UNIQUE INDEX "shard_instance_database_id_count_instance_database_shard_co_idx" ON "public"."shard_instance" USING btree ("database_id", "count", "instance", "database_shard", "collection_shard", "name");

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
ALTER TABLE "public"."collection_field" ADD FOREIGN KEY ("parent_collection_field_id") REFERENCES "public"."collection_field" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
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
