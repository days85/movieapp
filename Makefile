build_metadata:
	GOOS=linux go build -o metadata/main metadata/cmd/*.go
build_rating:
	GOOS=linux go build -o rating/main rating/cmd/*.go
build_movie:
	GOOS=linux go build -o movie/main movie/cmd/*.go
build:
	make build_metadata && make build_rating && make build_movie