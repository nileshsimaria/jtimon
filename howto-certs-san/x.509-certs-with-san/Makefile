all: ca-cert client-cert router-cert

ca-cert:
	openssl genrsa -out ca/ca.key 4096
	openssl req -x509 -new -key ca/ca.key \
		-days 3650 -out ca/ca.crt \
		-config ca/ca.cnf

client-cert:
	openssl genrsa -out client/client.key 4096
	openssl req -new -key client/client.key \
		-out client/client.csr -config client/client.cnf
	openssl x509 -req -days 365 -in client/client.csr \
		-CA ca/ca.crt -CAkey ca/ca.key \
		-CAcreateserial -out client/client.crt

router-cert:
	openssl genrsa -out router/router.key 4096
	openssl req -new -key router/router.key \
		-out router/router.csr -config router/router.cnf
	openssl x509 -req -days 365 -in router/router.csr \
		-CA ca/ca.crt -CAkey ca/ca.key \
		-CAcreateserial -out router/router.crt \
		-extensions req_ext -extfile router/router.cnf
	cat router/router.crt router/router.key > router/router.pem

read-ca-crt:
	openssl x509 -text -noout -in ca/ca.crt

read-client-csr:
	openssl req -noout -text -in client/client.csr

read-client-crt:
	openssl x509 -text -noout -in client/client.crt

read-router-csr:
	openssl req -noout -text -in router/router.csr

read-router-crt:
	openssl x509 -text -noout -in router/router.crt

clean:
	rm */*.crt */*.key */*.csr */*.srl */*.pem
