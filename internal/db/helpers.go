package db

import (
	"github.com/gocql/gocql"
)

func IsPartitionKeyColumn(col gocql.ColumnInfo, s *gocql.Session) bool {
	km, _ := s.KeyspaceMetadata(col.Keyspace)
	tm := km.Tables[col.Table]
	for _, c := range tm.PartitionKey {
		if c.Name == col.Name {
			return true
		}
	}
	return false
}

func IsClusterKeyColumn(col gocql.ColumnInfo, s *gocql.Session) bool {
	km, _ := s.KeyspaceMetadata(col.Keyspace)
	tm := km.Tables[col.Table]
	for _, c := range tm.ClusteringColumns {
		if c.Name == col.Name {
			return true
		}
	}
	return false
}

func IsStringColumn(col gocql.ColumnInfo) bool {
	t := col.TypeInfo.Type()
	switch t {
	case gocql.TypeAscii:
		return true
	case gocql.TypeText:
		return true
	case gocql.TypeVarchar:
		return true
	default:
		return false
	}
}
