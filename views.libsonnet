local sroto = import "sroto.libsonnet";
local droto = import "droto.jsonnet";

local isSerializer(x) =
    std.isObject(x) && std.objectHasAll(x, "droto_type")
    && x.droto_type == "serializer";

local isViewSet(x) =
    std.isObject(x) && std.objectHasAll(x, "droto_type")
    && x.droto_type == "viewset";

local serializerToViewMessage(serializer) =
    assert std.objectHas(serializer, "model");
    local model = serializer.model;
    local fields = {
        [n]: sroto.Field(model[n].getType(), model[n].number)
        for n in serializer.fields_arr
        if model[n].sroto_type == "field"
    } + std.get(serializer, "extra_fields", default={});
    local sname = serializer.name;
    local name = (
        if std.endsWith(sname, "Serializer")
        then std.substr(sname, 0, std.length(sname) - 10)
        else sname
    );
    sroto.Message(fields) {
        name: if name == model.name + "Response" then model.name else name,
        options+: model.options,
    };

local viewSetToService(viewset) =
    assert std.objectHas(viewset, "response_serializer");
    assert std.objectHas(viewset.response_serializer, "model");
    local model_name = viewset.response_serializer.model.name;
    local nameMapping = {
        list: "List" + model_name + "s",
        retrieve: "Get" + model_name,
    };
    local methodMapping = {
        list: sroto.UnaryMethod(std.substr(
            viewset.list_request_serializer.name, 0,
            std.length(viewset.list_request_serializer.name) - 10,
        ), viewset.list_response.name),
        retrieve: sroto.UnaryMethod(std.substr(
            viewset.get_request_serializer.name, 0,
            std.length(viewset.get_request_serializer.name) - 10,
        ), model_name),
    };
    sroto.Service({
        [nameMapping[c]]: methodMapping[c] for c in viewset.commands
    });

{
    File(name, package, file):: (
        local f1 = file {
            [(
                if std.endsWith(n, "ViewSet")
                then std.substr(n, 0, std.length(n) - 7)
                else n
            ) + "Service"]: viewSetToService(file[n])
            for n in std.objectFields(file)
            if isViewSet(file[n])
        };
        local f2 = f1 {
            [f1[n].get_request_serializer.name]: f1[n].get_request_serializer
            for n in std.objectFields(f1)
            if isViewSet(f1[n]) && !std.objectHas(f1, f1[n].get_request_serializer.name)
        } {
            [f1[n].list_request_serializer.name]: f1[n].list_request_serializer
            for n in std.objectFields(f1)
            if isViewSet(f1[n]) && !std.objectHas(f1, f1[n].list_request_serializer.name)
        } {
            [f1[n].list_response.name]: f1[n].list_response
            for n in std.objectFields(f1)
            if isViewSet(f1[n]) && !std.objectHas(f1, f1[n].list_request_serializer.name)
        };
        local f3 = f2 {
            [serializerToViewMessage({name: n} + f2[n]).name]:
            serializerToViewMessage({name: n} + f2[n])
            for n in std.objectFields(f2)
            if isSerializer(f2[n])
        };
        sroto.File(name, package, f3) {options+: [
            {
                type: droto.model_serializers,
                value: {
                    models_go_package: "github.com/tomlinford/droto/example/modelspb",
                    model_name: f3[n].model.name,
                    view_go_package: "github.com/tomlinford/droto/example/viewspb",
                    view_message_name: f3[n].model.name,
                    fields_arr: f3[n].fields_arr,
                },
            } for n in std.objectFields(f3) if isSerializer(f3[n])
        ]}
    ),
    RequestModelSerializer(model, fields_arr, extra_fields={}):: {
        droto_type:: "serializer",
        model: model,
        fields_arr: fields_arr,
        extra_fields: extra_fields,
    },
    ResponseModelSerializer(model, fields_arr):: {
        droto_type:: "serializer",
        model: model,
        fields_arr: fields_arr,
    },
    ReadOnlyModelViewSet(response_serializer):: {
        droto_type:: "viewset",
        local model = response_serializer.model,
        local default_get_request_fields = [
            f.name for f in model.getFields()
            if f.primary_key
        ],
        assert std.length(default_get_request_fields) == 1,
        get_request_serializer: $.RequestModelSerializer(
            response_serializer.model, default_get_request_fields,
        ) {name: "Get" + model.name + "RequestSerializer"},
        response_serializer: response_serializer,
        list_request_serializer: $.RequestModelSerializer(
            response_serializer.model, [], {
                cursor: sroto.StringField(100),
            },
        ) {name: "List" + model.name + "sRequestSerializer"},
        list_response: sroto.Message({
            results: sroto.Field(model.name, 1) {repeated: true},
            next: sroto.Field("List" + model.name + "sRequest", 2),
            prev: sroto.Field("List" + model.name + "sRequest", 3),
        }) {name: "List" + model.name + "sResponse"},
        commands: ["list", "retrieve"],
    },
}
