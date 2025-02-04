PROTO_DIR := .
PROTO_OUT := .
PROTOC := protoc

proto:
	@mkdir -p $(PROTO_OUT)
	$(PROTOC) \
		-I=$(PROTO_DIR) \
		--go_out=$(PROTO_OUT) \
		pb/api.proto