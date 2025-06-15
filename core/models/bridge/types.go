package bridge

type ExposeDataType = string
type ExposeCategory = string
type ExposeAccessMode = string

const (
	ReadAccessMode      ExposeAccessMode = "read"
	WriteAccessMode     ExposeAccessMode = "write"
	ReadWriteAccessMode ExposeAccessMode = "readwrite"
	UnknownAccessMode   ExposeAccessMode = "unknown"

	MeasurementCategory ExposeCategory = "measurement"
	DiagnosticCategory  ExposeCategory = "diagnostic"
	ConfigCategory      ExposeCategory = "config"

	NumericDataType   ExposeDataType = "numeric"
	EnumDataType      ExposeDataType = "enum"
	BinaryDataType    ExposeDataType = "binary"
	CompositeDataType ExposeDataType = "composite"
)
