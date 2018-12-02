package graph

type EditType int

const (
	Add EditType = iota
	Edit
	Delete
)

type NodeEdit struct {
	EditType EditType

	NodeIdx int64
	Data    interface{}
}

func NewNodeEdit(editType EditType, data interface{}, nodeIdx int64) NodeEdit {
	return NodeEdit{
		EditType: editType,
		Data:     data}
}

type ConnectionEdit struct {
	EditType EditType
	Data     interface{}

	ConnectionIdx  int64
	FirstNodeIdx   int64
	SecondNodeIdx  int64
	ConnectionData interface{}
}

func NewConnectionEdit(
	editType EditType,
	data interface{},
	connectionIdx, firstNodeIdx, secondNodeIdx int64,
	connectionData interface{}) ConnectionEdit {
	return ConnectionEdit{
		EditType:       editType,
		Data:           data,
		ConnectionIdx:  connectionIdx,
		FirstNodeIdx:   firstNodeIdx,
		SecondNodeIdx:  secondNodeIdx,
		ConnectionData: connectionData}
}
