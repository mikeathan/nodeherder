VERSION := 0.12.0

node-herder-deploy:	
	$(MAKE) all	
tag:
	echo "Creating tag..."
	git tag -a v$(VERSION) -m "Version $(VERSION)"
push_tag:
	echo "Pushing tag..."
	git push origin v$(VERSION)
all: tag push_tag