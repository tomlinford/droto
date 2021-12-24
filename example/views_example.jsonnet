local views = import "views.libsonnet";
local models = import "models_example.jsonnet";

views.File("viewspb/views.proto", "example.viewspb", {
    UserResponseSerializer: views.ResponseModelSerializer(models.User, [
        "id", "username", "about_me",
    ]),
    UserViewSet: views.ReadOnlyModelViewSet($.UserResponseSerializer),
}) {options+: [
    {"go_package": "github.com/tomlinford/droto/example/viewspb"},
]}
