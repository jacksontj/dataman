'''Example usage of dataman explained in a nice python script


The goal here is to showcase how a client application would use dataman.


For this example we'll be making a chat system (based on the tornado chat server demo)

'''
import argparse


import schema




if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument("--storage-node", required=True)

    args = parser.parse_args()

    # Create the database and tables
    schema.drop_db(args.storage_node)
    schema.create_db(args.storage_node)

    # Load some data


    # Create database / table
    # Load some data
    # Use it
    # Add a schema + Indexes
    # Load some data
    # Use it
    # update schema
    # Load some data
    # Use it
