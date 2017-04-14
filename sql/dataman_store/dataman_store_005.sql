/*
Navicat PGSQL Data Transfer

Source Server         : local
Source Server Version : 90506
Source Host           : localhost:5432
Source Database       : dataman_storagenode
Source Schema         : public

Target Server Type    : PGSQL
Target Server Version : 90506
File Encoding         : 65001

Date: 2017-04-13 12:48:01
*/


-- ----------------------------
-- Sequence structure for collection__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection__id_seq";
CREATE SEQUENCE "public"."collection__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for collection_field__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_field__id_seq";
CREATE SEQUENCE "public"."collection_field__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for collection_index__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_index__id_seq";
CREATE SEQUENCE "public"."collection_index__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for database__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database__id_seq";
CREATE SEQUENCE "public"."database__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;

-- ----------------------------
-- Sequence structure for schema__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."schema__id_seq";
CREATE SEQUENCE "public"."schema__id_seq"
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
"_id" int4 DEFAULT nextval('collection__id_seq'::regclass) NOT NULL,
"_created" timestamp(6),
"_updated" timestamp(6),
"id" int4,
"name" varchar(255) COLLATE "default",
"database_id" int4
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for collection_field
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_field";
CREATE TABLE "public"."collection_field" (
"_id" int4 DEFAULT nextval('collection_field__id_seq'::regclass) NOT NULL,
"_created" timestamp(6),
"_updated" timestamp(6),
"id" int4,
"name" varchar(255) COLLATE "default",
"collection_id" int4,
"field_type" varchar(255) COLLATE "default",
"order" int4,
"schema_id" int4,
"not_null" int4,
"field_type_args" jsonb
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for collection_index
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection_index";
CREATE TABLE "public"."collection_index" (
"_id" int4 DEFAULT nextval('collection_index__id_seq'::regclass) NOT NULL,
"_created" timestamp(6),
"_updated" timestamp(6),
"id" int4,
"name" varchar(255) COLLATE "default",
"collection_id" int4,
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
"_id" int4 DEFAULT nextval('database__id_seq'::regclass) NOT NULL,
"_created" timestamp(6),
"_updated" timestamp(6),
"id" int4,
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
"_created" timestamp(6),
"_updated" timestamp(6),
"id" int4,
"name" varchar(255) COLLATE "default",
"version" int4,
"data_json" jsonb,
"backwards_compatible" bool
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

-- ----------------------------
-- Indexes structure for table collection
-- ----------------------------
CREATE UNIQUE INDEX "index_collection_collection_name" ON "public"."collection" USING btree ("name", "database_id");
CREATE UNIQUE INDEX "index_collection_table_pkey" ON "public"."collection" USING btree ("id");

-- ----------------------------
-- Primary Key structure for table collection
-- ----------------------------
ALTER TABLE "public"."collection" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_field
-- ----------------------------
CREATE INDEX "index_collection_field_collection_field_table" ON "public"."collection_field" USING btree ("collection_id");
CREATE UNIQUE INDEX "index_collection_field_table_column_pkey" ON "public"."collection_field" USING btree ("id");
CREATE UNIQUE INDEX "index_collection_field_collection_field_name" ON "public"."collection_field" USING btree ("collection_id", "name");

-- ----------------------------
-- Primary Key structure for table collection_field
-- ----------------------------
ALTER TABLE "public"."collection_field" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_index
-- ----------------------------
CREATE UNIQUE INDEX "index_collection_index_collection_index_table" ON "public"."collection_index" USING btree ("name", "collection_id");
CREATE UNIQUE INDEX "index_collection_index_table_index_pkey" ON "public"."collection_index" USING btree ("id");

-- ----------------------------
-- Primary Key structure for table collection_index
-- ----------------------------
ALTER TABLE "public"."collection_index" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table database
-- ----------------------------
CREATE UNIQUE INDEX "index_database_name" ON "public"."database" USING btree ("name");
CREATE UNIQUE INDEX "index_database_database_pkey" ON "public"."database" USING btree ("id");

-- ----------------------------
-- Primary Key structure for table database
-- ----------------------------
ALTER TABLE "public"."database" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table schema
-- ----------------------------
CREATE UNIQUE INDEX "index_schema_schema_pkey" ON "public"."schema" USING btree ("id");
CREATE UNIQUE INDEX "index_schema_name_version" ON "public"."schema" USING btree ("name", "version");

-- ----------------------------
-- Primary Key structure for table schema
-- ----------------------------
ALTER TABLE "public"."schema" ADD PRIMARY KEY ("_id");
