VERSION := 0.5.5

node-herder-deploy: 

tag:
	echo "Creating tag..."
	git tag -a v$(VERSION) -m "Version $(VERSION)"

push_tag:
    echo "Pushing tag..."
    git push origin v$(VERSION)

all: tag push_tag
