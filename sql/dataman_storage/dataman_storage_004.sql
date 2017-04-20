/*
Navicat PGSQL Data Transfer

Source Server         : local
Source Server Version : 90506
Source Host           : localhost:5432
Source Database       : dataman_storage
Source Schema         : public

Target Server Type    : PGSQL
Target Server Version : 90506
File Encoding         : 65001

Date: 2017-04-12 11:36:35
*/


-- ----------------------------
-- Sequence structure for database_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database_id_seq";
CREATE SEQUENCE "public"."database_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 623
 CACHE 1;
SELECT setval('"public"."database_id_seq"', 623, true);

-- ----------------------------
-- Sequence structure for schema_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."schema_id_seq";
CREATE SEQUENCE "public"."schema_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 516
 CACHE 1;
SELECT setval('"public"."schema_id_seq"', 516, true);

-- ----------------------------
-- Sequence structure for table_column_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."table_column_id_seq";
CREATE SEQUENCE "public"."table_column_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 665
 CACHE 1;
SELECT setval('"public"."table_column_id_seq"', 665, true);

-- ----------------------------
-- Sequence structure for table_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."table_id_seq";
CREATE SEQUENCE "public"."table_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 874
 CACHE 1;
SELECT setval('"public"."table_id_seq"', 874, true);

-- ----------------------------
-- Sequence structure for table_index_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."table_index_id_seq";
CREATE SEQUENCE "public"."table_index_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 504
 CACHE 1;
SELECT setval('"public"."table_index_id_seq"', 504, true);

-- ----------------------------
-- Table structure for collection
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection";
CREATE TABLE "public"."collection" (
"id" int4 DEFAULT nextval('table_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"database_id" int4
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for collection_field
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_field";
CREATE TABLE "public"."collection_field" (
"id" int4 DEFAULT nextval('table_column_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"collection_id" int4 NOT NULL,
"field_type" varchar(255) COLLATE "default" NOT NULL,
"order" int4 NOT NULL,
"schema_id" int4,
"not_null" int2,
"field_type_args" jsonb
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for collection_index
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_index";
CREATE TABLE "public"."collection_index" (
"id" int4 DEFAULT nextval('table_index_id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default" NOT NULL,
"collection_id" int4 NOT NULL,
"data_json" jsonb,
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
"name" varchar(255) COLLATE "default" NOT NULL
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
-- Alter Sequences Owned By 
-- ----------------------------
ALTER SEQUENCE "public"."database_id_seq" OWNED BY "database"."id";
ALTER SEQUENCE "public"."schema_id_seq" OWNED BY "schema"."id";
ALTER SEQUENCE "public"."table_column_id_seq" OWNED BY "collection_field"."id";
ALTER SEQUENCE "public"."table_id_seq" OWNED BY "collection"."id";
ALTER SEQUENCE "public"."table_index_id_seq" OWNED BY "collection_index"."id";

-- ----------------------------
-- Indexes structure for table collection
-- ----------------------------
CREATE UNIQUE INDEX "collection_name" ON "public"."collection" USING btree ("name", "database_id");

-- ----------------------------
-- Primary Key structure for table collection
-- ----------------------------
ALTER TABLE "public"."collection" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table collection_field
-- ----------------------------
CREATE UNIQUE INDEX "collection_field_name" ON "public"."collection_field" USING btree ("collection_id", "name");
CREATE INDEX "collection_field_table" ON "public"."collection_field" USING btree ("collection_id");

-- ----------------------------
-- Primary Key structure for table collection_field
-- ----------------------------
ALTER TABLE "public"."collection_field" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table collection_index
-- ----------------------------
CREATE UNIQUE INDEX "collection_index_table" ON "public"."collection_index" USING btree ("name", "collection_id");

-- ----------------------------
-- Primary Key structure for table collection_index
-- ----------------------------
ALTER TABLE "public"."collection_index" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table database
-- ----------------------------
CREATE UNIQUE INDEX "name" ON "public"."database" USING btree ("name");

-- ----------------------------
-- Primary Key structure for table database
-- ----------------------------
ALTER TABLE "public"."database" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table schema
-- ----------------------------
CREATE UNIQUE INDEX "name_version" ON "public"."schema" USING btree ("name", "version");

-- ----------------------------
-- Primary Key structure for table schema
-- ----------------------------
ALTER TABLE "public"."schema" ADD PRIMARY KEY ("id");

-- ----------------------------
-- Foreign Key structure for table "public"."collection"
-- ----------------------------
ALTER TABLE "public"."collection" ADD FOREIGN KEY ("database_id") REFERENCES "public"."database" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_field"
-- ----------------------------
ALTER TABLE "public"."collection_field" ADD FOREIGN KEY ("collection_id") REFERENCES "public"."collection" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."collection_field" ADD FOREIGN KEY ("schema_id") REFERENCES "public"."schema" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_index"
-- ----------------------------
ALTER TABLE "public"."collection_index" ADD FOREIGN KEY ("collection_id") REFERENCES "public"."collection" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
