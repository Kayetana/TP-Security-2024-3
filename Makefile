start:
	$(CURDIR)/scripts/create_cert_and_keys.sh
	$(CURDIR)/scripts/build-proxy-image.sh
	$(CURDIR)/scripts/start-services.sh
