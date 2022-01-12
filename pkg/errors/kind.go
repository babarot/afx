package errors

// Kind defines the kind of error this is, mostly for use by systems
// such as FUSE that must act differently depending on the error.
type Kind uint8

// Kinds of errors.
//
// The values of the error kinds are common between both
// clients and servers. Do not reorder this list or remove
// any items since that will change their values.
// New items must be added only to the end.
const (
	Other         Kind = iota // Unclassified error. This value is not printed in the error message.
	Invalid                   // Invalid operation for this type of item.
	Permission                // Permission denied.
	IO                        // External I/O error such as network failure.
	Exist                     // Item already exists.
	NotExist                  // Item does not exist.
	IsDir                     // Item is a directory.
	NotDir                    // Item is not a directory.
	NotEmpty                  // Directory not empty.
	Private                   // Information withheld.
	Internal                  // Internal error or inconsistency.
	CannotDecrypt             // No wrapped key for user with read access.
	Transient                 // A transient error.
	BrokenLink                // Link target does not exist.
)

func (k Kind) String() string {
	switch k {
	case Other:
		return "other error"
	case Invalid:
		return "invalid operation"
	case Permission:
		return "permission denied"
	case IO:
		return "I/O error"
	case Exist:
		return "item already exists"
	case NotExist:
		return "item does not exist"
	case BrokenLink:
		return "link target does not exist"
	case IsDir:
		return "item is a directory"
	case NotDir:
		return "item is not a directory"
	case NotEmpty:
		return "directory not empty"
	case Private:
		return "information withheld"
	case Internal:
		return "internal error"
	case CannotDecrypt:
		return `no wrapped key for user; owner must "upspin share -fix"`
	case Transient:
		return "transient error"
	}
	return "unknown error kind"
}
