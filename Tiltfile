print("Wallet Files")

load("ext://restart_process", "docker_build_with_restart")

cfg = read_yaml(
    "tilt.yaml",
    default = read_yaml("tilt.yaml.sample"),
)

local_resource(
    "files-build-binary",
    "make fast_build",
    deps = ["./cmd", "./internal", "./rpc/pb_server.go"],
)
local_resource(
    "files-generate-protpbuf",
    "make gen-protobuf",
    deps = ["./rpc/files/files.proto"],
)

docker_build(
    "velmie/wallet-files-db-migration",
    ".",
    dockerfile = "Dockerfile.migrations",
    only = "migrations",
)
k8s_resource(
    "wallet-files-db-migration",
    trigger_mode = TRIGGER_MODE_MANUAL,
    resource_deps = ["wallet-files-db-init"],
)

wallet_files_options = dict(
    entrypoint = "/app/service_files",
    dockerfile = "Dockerfile.prebuild",
    port_forwards = [],
    helm_set = [],
)

if cfg["debug"]:
    wallet_files_options["entrypoint"] = "$GOPATH/bin/dlv --continue --listen :%s --accept-multiclient --api-version=2 --headless=true exec /app/service_files" % cfg["debug_port"]
    wallet_files_options["dockerfile"] = "Dockerfile.debug"
    wallet_files_options["port_forwards"] = cfg["debug_port"]
    wallet_files_options["helm_set"] = ["containerLivenessProbe.enabled=false", "containerPorts[0].containerPort=%s" % cfg["debug_port"]]

docker_build_with_restart(
    "velmie/wallet-files",
    ".",
    dockerfile = wallet_files_options["dockerfile"],
    entrypoint = wallet_files_options["entrypoint"],
    only = [
        "./build",
        "zoneinfo.zip",
    ],
    live_update = [
        sync("./build", "/app/"),
    ],
)
k8s_resource(
    "wallet-files",
    resource_deps = ["wallet-files-db-migration"],
    port_forwards = wallet_files_options["port_forwards"],
)

yaml = helm(
    "./helm/wallet-files",
    # The release name, equivalent to helm --name
    name = "wallet-files",
    # The values file to substitute into the chart.
    values = ["./helm/values-dev.yaml"],
    set = wallet_files_options["helm_set"],
)

k8s_yaml(yaml)
