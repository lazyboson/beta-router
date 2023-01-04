protoc -I "proto" --go_out=./ --go-grpc_out=./ --validate_out="lang=go:./" ./proto/api.proto
protoc -I "proto" --grpc-gateway_out=./ \
    --grpc-gateway_opt logtostderr=true \
    --grpc-gateway_opt paths=import \
    --grpc-gateway_opt generate_unbound_methods=true \
    ./proto/api.proto
protoc -I "proto" --openapiv2_out ./docs \
    --openapiv2_opt logtostderr=true \
    ./proto/api.proto
protoc --doc_out=./docs --doc_opt=html,index.html ./proto/api.proto
