Based on github.com/jlaffaye/ftp v0.2.4, retained MIT LICENSE and upstream tests.
Adds DialWithClearData to request PROT C while retaining implicit TLS control,
for paired A1/A1 mini firmware. Default upstream protected-data behavior remains.
The printer adapter pins control certificates and confines passive connections
to the same validated LAN IP. No credentials are emitted to debug logging.
