terraform {
  required_providers {
    k0s = {
      source = "tfplugin/k0s"
    }
  }
}

resource "k0s_cluster" "webzyno" {
  config = <<EOT
apiVersion: k0sctl.k0sproject.io/v1beta1
kind: Cluster
metadata:
  name: k0s
spec:
  hosts:
    - role: controller+worker
      ssh:
        address: 192.168.1.1
        keyPath: ./id_dsa
EOT
  ssh_private_key = <<EOT
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACBjg2i/oCrJOhfEcBRaL1LTRg8Ps/e66dLiaZ3c7HOXjgAAAJBXLL3lVyy9
5QAAAAtzc2gtZWQyNTUxOQAAACBjg2i/oCrJOhfEcBRaL1LTRg8Ps/e66dLiaZ3c7HOXjg
AAAEDfPefnvl0YaYhoJF2f0vjknRcAbMPN3RjAepZzsJLXImODaL+gKsk6F8RwFFovUtNG
Dw+z97rp0uJpndzsc5eOAAAABm5vbmFtZQECAwQFBgc=
-----END OPENSSH PRIVATE KEY-----

EOT
}
