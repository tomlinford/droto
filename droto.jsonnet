local sroto = import "sroto.libsonnet";

(sroto.File("droto.proto", "droto", {
    // nullable: sroto.CustomFieldOption("bool", 6061),
    field_options: sroto.CustomFieldOption("FieldOptions", 6061),
    FieldOptions: sroto.Message({
        nullable: sroto.Field("bool", 1),
        db_column: sroto.Field("string", 2),
    }),

    char_field_options: sroto.CustomFieldOption("CharFieldOptions", 6062),
    CharFieldOptions: sroto.Message({
        max_length: sroto.Field("int32", 1),
    }),

    ModelSerializer: sroto.Message({
        models_go_package: sroto.StringField(1),
        model_name: sroto.StringField(2),
        view_go_package: sroto.StringField(3),
        view_message_name: sroto.StringField(4),
        fields_arr: sroto.StringField(5) {repeated: true},
    }),
    model_serializers: sroto.CustomFileOption("ModelSerializer", 6063) {
        repeated: true,
    },
}) {options+: [
    {type: {name: "go_package"}, value: "github.com/tomlinford/droto"},
]})
