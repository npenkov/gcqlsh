package action

import (
	"fmt"

	"github.com/gocql/gocql"
)

func printRowValue(col gocql.ColumnInfo, value interface{}) string {
	typeMod := "s"
	t := col.TypeInfo.Type()
	switch t {
	case gocql.TypeCustom:
	case gocql.TypeAscii:
	case gocql.TypeBigInt:
		typeMod = "d"
	case gocql.TypeBlob:
	case gocql.TypeBoolean:
		typeMod = "t"
	case gocql.TypeCounter:
	case gocql.TypeDecimal:
		typeMod = "d"
	case gocql.TypeDouble:
		typeMod = "f"
	case gocql.TypeFloat:
		typeMod = "f"
	case gocql.TypeInt:
		typeMod = "d"
	case gocql.TypeText:
	case gocql.TypeTimestamp:
	case gocql.TypeUUID:
	case gocql.TypeVarchar:
	case gocql.TypeVarint:
	case gocql.TypeTimeUUID:
	case gocql.TypeInet:
	case gocql.TypeDate:
	case gocql.TypeTime:
	case gocql.TypeSmallInt:
		typeMod = "d"
	case gocql.TypeTinyInt:
		typeMod = "d"
	case gocql.TypeList:
	case gocql.TypeMap:
	case gocql.TypeSet:
	case gocql.TypeUDT:
	case gocql.TypeTuple:
	}
	val := fmt.Sprintf("%"+typeMod, value)
	return val
}
