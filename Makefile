.PHONY: protos

protos:
	protoc --twirp_out=. --go_out=. ./protos/api/api.proto
