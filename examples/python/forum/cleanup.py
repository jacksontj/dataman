import psycopg2
import psycopg2.extras

import requests

from schema import DBNAME

# Router node
if True:
    conn = psycopg2.connect("dbname=%s user='postgres' host='localhost' password='password'" % "dataman_router")

    cur = conn.cursor(cursor_factory=psycopg2.extras.DictCursor)


    queries = [
        "DELETE FROM collection_partition",
        "DELETE FROM collection_index_item",
        "DELETE FROM collection_index",
        "DELETE FROM collection_field_relation",
        "DELETE FROM collection_field",
        "DELETE FROM collection",
        "DELETE FROM datasource_instance_shard_instance",
        "DELETE FROM database_vshard_instance_datastore_shard",
        "DELETE FROM database_vshard_instance",
        "DELETE FROM database_vshard",
        "DELETE FROM database_datastore",
        "DELETE FROM database",
        "DELETE FROM datasource_instance_shard_instance"
    ]

    for q in queries:
        cur.execute(q)
        try:
            print cur.fetchall()
        except:
            pass

    conn.commit()
    cur.close()

# Storage node
if True:
    for addr in ('127.0.0.1', '10.42.17.93'):
        conn = psycopg2.connect("dbname=%s user='postgres' host='%s' password='password'" % ("dataman_storage", addr))
        conn.autocommit = True
        cur = conn.cursor(cursor_factory=psycopg2.extras.DictCursor)


        queries = [
            "DELETE FROM collection_index_item",
            "DELETE FROM collection_index",
            "DELETE FROM collection_field_relation",
            "DELETE FROM collection_field",
            "DELETE FROM collection",
            "DELETE FROM shard_instance",
            "DELETE FROM database",
        ]

        for q in queries:
            cur.execute(q)
            try:
                print cur.fetchall()
            except:
                pass
        conn.close()

        conn = psycopg2.connect("user='postgres' host='%s' password='password'" % addr)
        conn.autocommit = True
        cur = conn.cursor(cursor_factory=psycopg2.extras.DictCursor)

        tasks = [
            # Kick everyone off
            'SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = \'%s\' AND pid <> pg_backend_pid();' % DBNAME,
            # Drop it
            'DROP DATABASE IF EXISTS "%s"' % DBNAME,
        ]
        
        for task in tasks:
                cur.execute(task)

if True:
    for addr in ('127.0.0.1', '10.42.17.93'):
        requests.delete('http://'+addr+':8081/v1/datasource_instance/postgres1/database/example_forum')

            
