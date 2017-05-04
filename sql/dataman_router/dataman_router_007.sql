/*
Navicat PGSQL Data Transfer

Source Server         : local
Source Server Version : 90602
Source Host           : localhost:5432
Source Database       : dataman_router
Source Schema         : public

Target Server Type    : PGSQL
Target Server Version : 90602
File Encoding         : 65001

Date: 2017-05-04 13:26:58
*/


-- ----------------------------
-- Sequence structure for collection__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection__id_seq";
CREATE SEQUENCE "public"."collection__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 184
 CACHE 1;
SELECT setval('"public"."collection__id_seq"', 184, true);

-- ----------------------------
-- Sequence structure for collection_field__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_field__id_seq";
CREATE SEQUENCE "public"."collection_field__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 184
 CACHE 1;
SELECT setval('"public"."collection_field__id_seq"', 184, true);

-- ----------------------------
-- Sequence structure for collection_index__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_index__id_seq";
CREATE SEQUENCE "public"."collection_index__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 136
 CACHE 1;
SELECT setval('"public"."collection_index__id_seq"', 136, true);

-- ----------------------------
-- Sequence structure for collection_partition_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."collection_partition_id_seq";
CREATE SEQUENCE "public"."collection_partition_id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 180
 CACHE 1;
SELECT setval('"public"."collection_partition_id_seq"', 180, true);

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
-- Sequence structure for database__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database__id_seq";
CREATE SEQUENCE "public"."database__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 72
 CACHE 1;
SELECT setval('"public"."database__id_seq"', 72, true);

-- ----------------------------
-- Sequence structure for database_datastore__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database_datastore__id_seq";
CREATE SEQUENCE "public"."database_datastore__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 57
 CACHE 1;
SELECT setval('"public"."database_datastore__id_seq"', 57, true);

-- ----------------------------
-- Sequence structure for database_vshard__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database_vshard__id_seq";
CREATE SEQUENCE "public"."database_vshard__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 62
 CACHE 1;
SELECT setval('"public"."database_vshard__id_seq"', 62, true);

-- ----------------------------
-- Sequence structure for database_vshard_instance__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database_vshard_instance__id_seq";
CREATE SEQUENCE "public"."database_vshard_instance__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 267
 CACHE 1;
SELECT setval('"public"."database_vshard_instance__id_seq"', 267, true);

-- ----------------------------
-- Sequence structure for database_vshard_instance_datastore_shard__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."database_vshard_instance_datastore_shard__id_seq";
CREATE SEQUENCE "public"."database_vshard_instance_datastore_shard__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 264
 CACHE 1;
SELECT setval('"public"."database_vshard_instance_datastore_shard__id_seq"', 264, true);

-- ----------------------------
-- Sequence structure for datasource__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datasource__id_seq";
CREATE SEQUENCE "public"."datasource__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 1
 CACHE 1;
SELECT setval('"public"."datasource__id_seq"', 1, true);

-- ----------------------------
-- Sequence structure for datasource_instance__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datasource_instance__id_seq";
CREATE SEQUENCE "public"."datasource_instance__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 2
 CACHE 1;
SELECT setval('"public"."datasource_instance__id_seq"', 2, true);

-- ----------------------------
-- Sequence structure for datasource_instance_shard_instance__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datasource_instance_shard_instance__id_seq";
CREATE SEQUENCE "public"."datasource_instance_shard_instance__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 236
 CACHE 1;
SELECT setval('"public"."datasource_instance_shard_instance__id_seq"', 236, true);

-- ----------------------------
-- Sequence structure for datastore__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datastore__id_seq";
CREATE SEQUENCE "public"."datastore__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 2
 CACHE 1;
SELECT setval('"public"."datastore__id_seq"', 2, true);

-- ----------------------------
-- Sequence structure for datastore_shard__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datastore_shard__id_seq";
CREATE SEQUENCE "public"."datastore_shard__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 4
 CACHE 1;
SELECT setval('"public"."datastore_shard__id_seq"', 4, true);

-- ----------------------------
-- Sequence structure for datastore_shard_replica__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."datastore_shard_replica__id_seq";
CREATE SEQUENCE "public"."datastore_shard_replica__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 5
 CACHE 1;
SELECT setval('"public"."datastore_shard_replica__id_seq"', 5, true);

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
-- Sequence structure for storage_node_type__id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."storage_node_type__id_seq";
CREATE SEQUENCE "public"."storage_node_type__id_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 7
 CACHE 1;
SELECT setval('"public"."storage_node_type__id_seq"', 7, true);

-- ----------------------------
-- Table structure for collection
-- ----------------------------
DROP TABLE IF EXISTS "public"."collection";
CREATE TABLE "public"."collection" (
"_id" int4 DEFAULT nextval('collection__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"database_id" int4,
"collection_vshard_id" int4
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
"field_type_args" jsonb,
"not_null" bool,
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
-- Table structure for database_datastore
-- ----------------------------
DROP TABLE IF EXISTS "public"."database_datastore";
CREATE TABLE "public"."database_datastore" (
"_id" int4 DEFAULT nextval('database_datastore__id_seq'::regclass) NOT NULL,
"database_id" int4,
"datastore_id" int4,
"read" bool,
"write" bool,
"required" bool
)
WITH (OIDS=FALSE)

;

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
-- Table structure for database_vshard_instance
-- ----------------------------
DROP TABLE IF EXISTS "public"."database_vshard_instance";
CREATE TABLE "public"."database_vshard_instance" (
"_id" int4 DEFAULT nextval('database_vshard_instance__id_seq'::regclass) NOT NULL,
"database_vshard_id" int4,
"shard_instance" int4 NOT NULL
)
WITH (OIDS=FALSE)

;

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
-- Table structure for datasource
-- ----------------------------
DROP TABLE IF EXISTS "public"."datasource";
CREATE TABLE "public"."datasource" (
"_id" int4 DEFAULT nextval('datasource__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"config_json_schema_id" int4
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
"config_json" jsonb
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
"database_vshard_instance_id" int4,
"collection_vshard_instance_id" int4,
"name" varchar(255) COLLATE "default"
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for datastore
-- ----------------------------
DROP TABLE IF EXISTS "public"."datastore";
CREATE TABLE "public"."datastore" (
"_id" int4 DEFAULT nextval('datastore__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default"
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
"shard_instance" int4 NOT NULL
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
"master" bool NOT NULL
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Table structure for schema
-- ----------------------------
DROP TABLE IF EXISTS "public"."schema";
CREATE TABLE "public"."schema" (
"_id" int4 DEFAULT nextval('schema__id_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"version" int4,
"data_json" jsonb,
"backwards_compatible" bool
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
"port" int4
)
WITH (OIDS=FALSE)

;

-- ----------------------------
-- Alter Sequences Owned By 
-- ----------------------------
ALTER SEQUENCE "public"."collection__id_seq" OWNED BY "collection"."_id";
ALTER SEQUENCE "public"."collection_field__id_seq" OWNED BY "collection_field"."_id";
ALTER SEQUENCE "public"."collection_index__id_seq" OWNED BY "collection_index"."_id";
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
ALTER SEQUENCE "public"."schema__id_seq" OWNED BY "schema"."_id";
ALTER SEQUENCE "public"."storage_node_type__id_seq" OWNED BY "storage_node"."_id";

-- ----------------------------
-- Indexes structure for table collection
-- ----------------------------
CREATE UNIQUE INDEX "index_index_collection_collection_name" ON "public"."collection" USING btree ("name", "database_id");
CREATE UNIQUE INDEX "index_collection_pkey" ON "public"."collection" USING btree ("_id");

-- ----------------------------
-- Primary Key structure for table collection
-- ----------------------------
ALTER TABLE "public"."collection" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_field
-- ----------------------------
CREATE INDEX "index_index_collection_field_collection_field_table" ON "public"."collection_field" USING btree ("collection_id");
CREATE UNIQUE INDEX "index_index_collection_field_collection_field_name" ON "public"."collection_field" USING btree ("collection_id", "name");
CREATE UNIQUE INDEX "index_collection_field_pkey" ON "public"."collection_field" USING btree ("_id");

-- ----------------------------
-- Primary Key structure for table collection_field
-- ----------------------------
ALTER TABLE "public"."collection_field" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table collection_index
-- ----------------------------
CREATE UNIQUE INDEX "index_collection_index_pkey" ON "public"."collection_index" USING btree ("_id");
CREATE UNIQUE INDEX "index_collection_index_name" ON "public"."collection_index" USING btree ("name", "collection_id");

-- ----------------------------
-- Primary Key structure for table collection_index
-- ----------------------------
ALTER TABLE "public"."collection_index" ADD PRIMARY KEY ("_id");

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
CREATE UNIQUE INDEX "index_database_pkey" ON "public"."database" USING btree ("_id");

-- ----------------------------
-- Primary Key structure for table database
-- ----------------------------
ALTER TABLE "public"."database" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table database_datastore
-- ----------------------------
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
-- Primary Key structure for table datastore
-- ----------------------------
ALTER TABLE "public"."datastore" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table datastore_shard
-- ----------------------------
CREATE UNIQUE INDEX "datastore_shard_number" ON "public"."datastore_shard" USING btree ("datastore_id", "shard_instance");
CREATE UNIQUE INDEX "datastore_shard_name_datastore_id_idx" ON "public"."datastore_shard" USING btree ("name", "datastore_id");

-- ----------------------------
-- Primary Key structure for table datastore_shard
-- ----------------------------
ALTER TABLE "public"."datastore_shard" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Primary Key structure for table datastore_shard_replica
-- ----------------------------
ALTER TABLE "public"."datastore_shard_replica" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table schema
-- ----------------------------
CREATE UNIQUE INDEX "index_index_schema_name_version" ON "public"."schema" USING btree ("name", "version");
CREATE UNIQUE INDEX "index_schema_pkey" ON "public"."schema" USING btree ("_id");

-- ----------------------------
-- Primary Key structure for table schema
-- ----------------------------
ALTER TABLE "public"."schema" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Indexes structure for table storage_node
-- ----------------------------
CREATE UNIQUE INDEX "storage_node_name_idx" ON "public"."storage_node" USING btree ("name");

-- ----------------------------
-- Primary Key structure for table storage_node
-- ----------------------------
ALTER TABLE "public"."storage_node" ADD PRIMARY KEY ("_id");

-- ----------------------------
-- Foreign Key structure for table "public"."collection"
-- ----------------------------
ALTER TABLE "public"."collection" ADD FOREIGN KEY ("database_id") REFERENCES "public"."database" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."collection" ADD FOREIGN KEY ("collection_vshard_id") REFERENCES "public"."collection_vshard" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_field"
-- ----------------------------
ALTER TABLE "public"."collection_field" ADD FOREIGN KEY ("parent_collection_field_id") REFERENCES "public"."collection_field" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."collection_field" ADD FOREIGN KEY ("collection_id") REFERENCES "public"."collection" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."collection_index"
-- ----------------------------
ALTER TABLE "public"."collection_index" ADD FOREIGN KEY ("collection_id") REFERENCES "public"."collection" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

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
ALTER TABLE "public"."collection_vshard_instance_datastore_shard" ADD FOREIGN KEY ("datastore_shard_id") REFERENCES "public"."datastore_shard" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."collection_vshard_instance_datastore_shard" ADD FOREIGN KEY ("collection_vshard_instance_id") REFERENCES "public"."collection_vshard_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

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
ALTER TABLE "public"."database_vshard_instance_datastore_shard" ADD FOREIGN KEY ("datastore_shard_id") REFERENCES "public"."datastore_shard" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."database_vshard_instance_datastore_shard" ADD FOREIGN KEY ("database_vshard_instance_id") REFERENCES "public"."database_vshard_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datasource"
-- ----------------------------
ALTER TABLE "public"."datasource" ADD FOREIGN KEY ("config_json_schema_id") REFERENCES "public"."schema" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datasource_instance"
-- ----------------------------
ALTER TABLE "public"."datasource_instance" ADD FOREIGN KEY ("storage_node_id") REFERENCES "public"."storage_node" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."datasource_instance" ADD FOREIGN KEY ("datasource_id") REFERENCES "public"."datasource" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datasource_instance_shard_instance"
-- ----------------------------
ALTER TABLE "public"."datasource_instance_shard_instance" ADD FOREIGN KEY ("datasource_instance_id") REFERENCES "public"."datasource_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."datasource_instance_shard_instance" ADD FOREIGN KEY ("database_vshard_instance_id") REFERENCES "public"."database_vshard_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."datasource_instance_shard_instance" ADD FOREIGN KEY ("collection_vshard_instance_id") REFERENCES "public"."collection_vshard_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datastore_shard"
-- ----------------------------
ALTER TABLE "public"."datastore_shard" ADD FOREIGN KEY ("datastore_id") REFERENCES "public"."datastore" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Key structure for table "public"."datastore_shard_replica"
-- ----------------------------
ALTER TABLE "public"."datastore_shard_replica" ADD FOREIGN KEY ("datastore_shard_id") REFERENCES "public"."datastore_shard" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."datastore_shard_replica" ADD FOREIGN KEY ("datasource_instance_id") REFERENCES "public"."datasource_instance" ("_id") ON DELETE NO ACTION ON UPDATE NO ACTION;
