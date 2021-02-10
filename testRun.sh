go run cmd/main.go --path=/tmp/node1 --grpcip=127.0.0.1 --grpc-port 19101 --ml-bind-addr=127.0.0.1 --ml-bind-port=8001 &

sleep 1

go run cmd/main.go --path=/tmp/node2 --grpcip=127.0.0.2 --grpc-port 19101  --ml-bind-addr=127.0.0.2 --ml-bind-port=8001 --ml-members=127.0.0.1:8001


# m@dev:~/Work/grkv$ go run cmd/main.go -h
# Usage: main [options] [arguments]

# OPTIONS
#   --path/$RKV_PATH                                    <string>    (default: /tmp/rkv)
#   --no-sync/$RKV_NO_SYNC                              <bool>      (default: false)
#   --value-log-gc/$RKV_VALUE_LOG_GC                    <bool>      (default: true)
#   --gc-interval/$RKV_GC_INTERVAL                      <duration>  (default: 1m)
#   --mandatory-gc-interval/$RKV_MANDATORY_GC_INTERVAL  <duration>  (default: 10m)
#   --gc-threshold/$RKV_GC_THRESHOLD                    <int>       (default: 1000000)
#   --grpcip/$RKV_GRPCIP                                <string>    (default: 127.0.0.1)
#   --grpc-port/$RKV_GRPC_PORT                          <int>       (default: 9001)
#   --ml-bind-addr/$RKV_ML_BIND_ADDR                    <string>    (default: 127.0.0.1)
#   --ml-bind-port/$RKV_ML_BIND_PORT                    <int>       (default: 8001)
#   --ml-members/$RKV_ML_MEMBERS                        <string>
#   --help/-h
#   display this help message