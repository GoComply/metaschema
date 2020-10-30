package parser

type AsType string

const (
	AsTypeBoolean            AsType = "boolean"
	AsTypeEmpty              AsType = "empty"
	AsTypeString             AsType = "string"
	AsTypeMixed              AsType = "mixed"
	AsTypeMarkupLine         AsType = "markup-line"
	AsTypeMarkupMultiLine    AsType = "markup-multiline"
	AsTypeDate               AsType = "date"
	AsTypeDateTimeTZ         AsType = "dateTime-with-timezone"
	AsTypeNCName             AsType = "NCName"
	AsTypeEmail              AsType = "email"
	AsTypeURI                AsType = "uri"
	AsTypeBase64             AsType = "base64Binary"
	AsTypeIDRef              AsType = "IDREF"
	AsTypeNMToken            AsType = "NMTOKEN"
	AsTypeID                 AsType = "ID"
	AsTypeAnyURI             AsType = "anyURI"
	AsTypeURIRef             AsType = "uri-reference"
	AsTypeUUID               AsType = "uuid"
	AsTypeNonNegativeInteger AsType = "nonNegativeInteger"
)

var goDatatypeMap = map[AsType]string{
	AsTypeString:             "string",
	AsTypeIDRef:              "string",
	AsTypeNCName:             "string",
	AsTypeNMToken:            "string",
	AsTypeID:                 "string",
	AsTypeURIRef:             "string",
	AsTypeURI:                "string",
	AsTypeUUID:               "string",
	AsTypeNonNegativeInteger: "uint64",
}
