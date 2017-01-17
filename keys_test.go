package envconfig

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeys(t *testing.T) {
	name := fieldName{"CassandraSslCert"}
	keys := name.Keys()

	require.Equal(t, 4, len(keys))
	require.Equal(t, "CASSANDRASSLCERT", keys[0])
	require.Equal(t, "CASSANDRA_SSL_CERT", keys[1])
	require.Equal(t, "cassandra_ssl_cert", keys[2])
	require.Equal(t, "cassandrasslcert", keys[3])

	name = fieldName{"CassandraSSLCert"}
	keys = name.Keys()

	require.Equal(t, 4, len(keys))
	require.Equal(t, "CASSANDRASSLCERT", keys[0])
	require.Equal(t, "CASSANDRA_SSL_CERT", keys[1])
	require.Equal(t, "cassandra_ssl_cert", keys[2])
	require.Equal(t, "cassandrasslcert", keys[3])

	name = fieldName{"Cassandra", "SslCert"}
	keys = name.Keys()

	require.Equal(t, 4, len(keys))
	require.Equal(t, "CASSANDRA_SSLCERT", keys[0])
	require.Equal(t, "CASSANDRA_SSL_CERT", keys[1])
	require.Equal(t, "cassandra_ssl_cert", keys[2])
	require.Equal(t, "cassandra_sslcert", keys[3])

	name = fieldName{"Cassandra", "SSLCert"}
	keys = name.Keys()

	require.Equal(t, 4, len(keys))
	require.Equal(t, "CASSANDRA_SSLCERT", keys[0])
	require.Equal(t, "CASSANDRA_SSL_CERT", keys[1])
	require.Equal(t, "cassandra_ssl_cert", keys[2])
	require.Equal(t, "cassandra_sslcert", keys[3])

	name = fieldName{"Name"}
	keys = name.Keys()

	require.Equal(t, 2, len(keys))
	require.Equal(t, "NAME", keys[0])
	require.Equal(t, "name", keys[1])
}
