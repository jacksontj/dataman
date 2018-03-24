package metadata

// TODO: put into database as a table

type ProvisionState int

const (
	Config ProvisionState = iota
	Provision
	Validate
	Active
	Maintenance

	// TODO: do we need this? If we don't then we need to have a separate mechanism
	// to know when something is on the way out
	Deallocate
)
