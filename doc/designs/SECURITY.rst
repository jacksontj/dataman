# Security


## Access control
The end-goal here is to have a very pluggable authentication and authorization system. Initially we'll have
just user/password with controls on database/table/record.

The ideal end-state will support:
    - client certs
    - revokable session tokens
        -- keep track of when/who issued them
        -- current use
        -- ability to revoke
    - per-table access controls
    - per-record access controls



## Encryption

### In-Flight
In flight we'll be using TLS to encrypt traffic between all layers in the middle (ideally with client certs).



### At-Rest
Optional encrption of data before storing in the storage node (meaning dataman has to decrypt the entry
before returning)

TODO: This will present some interesting issues-- as indexes are either not-encrypted, or non-existant





QUESTIONS:
- How do we want to do login?
    -- We'd like to give out session tokens (which would be tied to a particular host or
        session, so we can track who is acccessing what from where
    -- Login Options:
        -- separate "login" endpoint that gives back tokens (probably this)
            -- not sure how that would tie into client certs? Maybe the client cert is
                just part of the "key" for the token, meaning they still have to sign-in
                and if the client-cert changes or is revoked we'll kill the session token
        -- HTTP Basic auth
            -- I don't like sending the credentials around so much :/
            -- Not a session, meaning we have to revalidate creds every time
