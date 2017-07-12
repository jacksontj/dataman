package functiondefault

type FunctionDefaultType string

const (
	UUID4Type  FunctionDefaultType = "uuid4"
	RandomType                     = "random"
	KSUIDType                      = "ksuid"
)

func (s FunctionDefaultType) Get() FunctionDefault {
	switch s {
	case UUID4Type:
		return &UUID4{}
	case RandomType:
		return &Random{}
	case KSUIDType:
		return &KSUID{}
	default:
		return nil
	}
}
