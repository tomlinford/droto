local sroto = import "sroto.libsonnet";
local models = import "models.libsonnet";

sroto.File("modelspb/models.proto", "example.modelspb", {
    User: models.Model({
        id: models.BigIntegerField(1) {primary_key: true},
        username: models.CharField(2, 256),
        about_me: models.CharField(3, 1024) {nullable: true},
        password_hash: models.CharField(4, 256),
        password_salt: models.CharField(5, 256),
    })
}) {options+: [
    {"go_package": "github.com/tomlinford/droto/example/modelspb"},
]}
