# Usage:
# make        # creates test.key and test.cert files


.PHONY: all gencert

all: gencert

gencert:
	@echo "generating self signed cert.."
	openssl req \
		-newkey rsa:2048 -nodes -keyout test.key \
		-subj '/C=XX/ST=XX/L=XX/O=XX/CN=example.com' \
		-x509 -days 365 -out test.crt

readcert:
	@echo "reading generated cert.."
	openssl x509 -text -noout -in test.crt
