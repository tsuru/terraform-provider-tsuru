resource "tsuru_cluster" "k8s_cluster" {
  name        = "my-k8s-cluster"
  provisioner = "kubernetes"
  crio        = true

  kube_config {
    cluster {
      server = "https://my-k8s-cluster.example.com"
    }
  }
}

resource "tsuru_pool" "my_pool" {
  name        = "my-pool"
  provisioner = "kubernetes"
}

resource "tsuru_cluster_pool" "cluster_pool_attachment" {
  cluster = tsuru_cluster.k8s_cluster.name
  pool    = tsuru_pool.my_pool.name
}

