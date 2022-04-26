package debug

const (
	// LogDNAIngestionKeyName is the name of the LogDNA ingestion key in the keyfile
	LogDNAIngestionKeyName = "logdna-osscat-ingestion"

	// LogDNAIngestionURL is the URL for sending logs to be ingested by LogDNA
	LogDNAIngestionURL = "https://logs.us-south.logging.cloud.ibm.com/logs/ingest"

	// LogDNAServiceKeyName is the name of the LogDNA service key in the keyfile
	LogDNAServiceKeyName = "logdna-osscat-service"

	// LogDNAExportURL is the URL for exporting logs from LogDNA
	LogDNAExportURL = "https://api.us-south.logging.cloud.ibm.com/v1/export"
)
