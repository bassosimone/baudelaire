.PHONY: clean install

BAUDELAIRE = baudelaire-linux-amd64
DEPLOY_HOST = # To be set from the command line

$(BAUDELAIRE): main.go
	GOARCH=amd64 GOOS=linux go build -v -o $(BAUDELAIRE)

clean:
	rm -rf -- $(BAUDELAIRE) baudelaire

deploy: $(BAUDELAIRE) baudelaire.service deploy.sh rc.local
	if test -z "$(DEPLOY_HOST)"; then                                      \
	  echo "usage: make deploy DEPLOY_HOST=1.2.3.4" 1>&2;                  \
	  exit 1;                                                              \
	fi
	scp $(BAUDELAIRE) baudelaire.service deploy.sh rc.local $(DEPLOY_HOST):
