package db

import (
	"time"

	"github.com/gocql/gocql"
)

func createCluster(host string, port int, keyspace string) *gocql.ClusterConfig {
	cluster := gocql.NewCluster(gocql.JoinHostPort(host, port))

	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.One
	cluster.Timeout = 10 * time.Second
	cluster.MaxWaitSchemaAgreement = 2 * time.Minute
	cluster.ProtoVersion = 3
	cluster.IgnorePeerAddr = true
	cluster.DisableInitialHostLookup = true

	cluster.NumConns = 3

	return cluster
}

func createSession(cluster *gocql.ClusterConfig) (*gocql.Session, func(), error) {
	session, err := cluster.CreateSession()
	return session, func() {
		session.Close()
	}, err
}

func NewSession(host string, port int, keyspace string) (*gocql.Session, func(), error) {
	return createSession(createCluster(host, port, keyspace))
}
