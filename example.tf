terraform {
    required_providers {
        sakuracloud = {
            source  = "sacloud/sakuracloud"
            version = "2.25.0"
        }
    }
}
provider "sakuracloud" {
    profile = "default"
}

resource "sakuracloud_server" "test-server" {
    name = "test-server"
    core = 1
    memory = 1
    disks = ["113600185504"]
}