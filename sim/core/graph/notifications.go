package graph

type EditType int

const (
	Add EditType = iota
	Edit
	Delete
)

type NodeEdit struct {
	EditType EditType

	NodeIdx int
	Data    interface{}
}

func NewNodeEdit(editType EditType, data interface{}, nodeIdx int) NodeEdit {
	return NodeEdit{
		EditType: editType,
		Data:     data}
}

type ConnectionEdit struct {
	EditType EditType
	Data     interface{}

	FirstNodeIdx   int
	SecondNodeIdx  int
	ConnectionData interface{}
}

func NewConnectionEdit(
	editType EditType,
	data interface{},
	firstNodeIdx, secondNodeIdx int,
	connectionData interface{}) ConnectionEdit {
	return ConnectionEdit{
		EditType:       editType,
		Data:           data,
		FirstNodeIdx:   firstNodeIdx,
		SecondNodeIdx:  secondNodeIdx,
		ConnectionData: connectionData}
}
