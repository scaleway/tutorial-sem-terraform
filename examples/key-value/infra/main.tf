locals {
  # get json
  sensitive_content = sensitive(file("${path.module}/sensitive.json"))
}

# SCALEWAY SECRET
resource "scaleway_secret" "main" {
  name        = "database_secret"
  description = "This is a key/value secret"
  tags        = ["devtools"]
}

# SCALEWAY SECRET VERSION
resource "scaleway_secret_version" "v1" {
  description = "version1"
  secret_id   = scaleway_secret.main.id
  data        = local.sensitive_content
}

// Get Secret version from a json Object
data "scaleway_secret_version" "secret" {
  secret_name = "database_secret"
  revision    = "1"
  depends_on  = [scaleway_secret.main, scaleway_secret_version.v1]
}

locals {
  db_cred = jsondecode(base64decode(data.scaleway_secret_version.secret.data))
}

// Create Database instance with credentials from SEM
resource "scaleway_rdb_instance" "pgsql" {
  name           = "dataBaseInstance"
  node_type      = "db-dev-m"
  engine         = "PostgreSQL-13"
  is_ha_cluster  = false
  disable_backup = true
  user_name      = local.db_cred.username
  password       = local.db_cred.password
}

output "database_public_endpoint" {
  value = scaleway_rdb_instance.pgsql.load_balancer
}
