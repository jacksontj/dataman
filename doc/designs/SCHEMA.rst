# Schema

We want to optionally support schema enforcement for records in tables.


We'll be using JSONSchema (http://json-schema.org/) as the definition language, here is an example from (http://json-schema.org/examples.html):
    {
	    "title": "Person",
	    "type": "object",
	    "properties": {
		    "firstName": {
			    "type": "string"
		    },
		    "lastName": {
			    "type": "string"
		    },
		    "age": {
			    "description": "Age in years",
			    "type": "integer",
			    "minimum": 0
		    }
	    },
	    "required": ["firstName", "lastName"]
    }

With this format we can define what the object is, and the fields that must be there. JSONSchema also allows
for sub-schemas embedded within the objects (referenced by URI to the schema).

In addition JSONSchema allows us to define "defaults" for properties, which will allow us to add new fields
with defaults-- this way we can update the rows with those values (so the clients don't need to handle
these fields existing or not existing).


We should also allow for different schema enforcement modes:
    - NONE: free-for-all (k/v store)
    - WARNING: helpful for ramping
    - SUBSET:  meaning you can have *more* if you want
    - STRICT: meaning it must match exactly


TODO: In addition to simply defining the record schema, we need to define what indexes will be present for the table.


TODOC:
    - We will *not* allow for nested documents, if you want nested documents you can add the child schema
        as a property of the schema .
    - When do we want to enforce schema?
        -- On-write
            pros:
                - cheaper (only done on a change)
            cons:
                - new fields will be empty until the objects are re-written
        -- On-Read
            pros:
                - consistent object view on retrieval
            cons:
                - expensive (we have to validate each object all the time)
        -- On-Demand (middle-ground) -- THIS ONE
            If each document stores a "schema version" (just the number) then we can
            enforce the new schema only when it is accessed (read or write)
    - Should we allow multiple document schemas in a given table?
        -- No is *much* easier, for indexes etc. -- NO


Notes:
    - helpful editor -- https://jsonschemalint.com/#/version/draft-04/markup/json
    - book on jsonschema: https://spacetelescope.github.io/understanding-json-schema/index.html

Implementation Notes:
    // Here is a golang one that does defaults etc.
    "github.com/lestrrat/go-jsschema"
	"github.com/lestrrat/go-jsval/builder"


Questions:
    - How do we want to handle schema changes?
        - rollbacks?
            -- since we are going to keep track of the schemas as versions, this should be fine.
                we just need to decide what to do when a field goes away -- wheter we leave it or drop it
                    -- we can either always do something
                        -- always leave it, require explicit action to "trim" data
                    -- have the action depend on schema enforcement mode
        - how do we want to approve them (for starters, probably approved by admin)
        -
