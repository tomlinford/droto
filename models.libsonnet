local sroto = import "sroto.libsonnet";
local droto = import "droto.jsonnet";

{
    Field(number):: sroto.LazilyTypedField(number) {
        nullable: false,
        blank: false,
        choices: null,
        db_column: "",
        db_index: false,
        db_tablespace: null,
        //default: ?
        editable: true,
        error_messages: null,
        //help_text: ?
        primary_key: false,
        unique: false,
        unique_for_date: false,
        unique_for_month: false,
        unique_for_year: false,
        //verbose_name: ?
        //validators: ?

        getOptions():: local f = self; super.getOptions() + [{
            type: droto.field_options,
            value: {
                [if f.nullable then "nullable" else null]: f.nullable,
                [if (
                    f.db_column != "" && f.db_column != f.name
                ) then "db_column" else null]: f.db_column,
            },
        }],
    },
    CharField(number, max_length):: self.Field(number) {
        max_length: max_length,

        getType():: if self.nullable then sroto.WKT.StringValue else "string",
        getOptions():: local f = self; super.getOptions() + [{
            type: droto.char_field_options,
            value: {
                max_length: f.max_length,
            },
        }],
    },
    BigIntegerField(number):: self.Field(number) {
        getType():: if self.nullable then sroto.WKT.Int64Value else "int64",
    },

    Model(fields):: sroto.Message(fields) {
        getFields():: local m = self; [
            m[n] for n in std.objectFields(m)
            if std.isObject(m[n]) && std.objectHasAll(m[n], "sroto_type")
            && m[n].sroto_type == "field"
        ],
    },
}
