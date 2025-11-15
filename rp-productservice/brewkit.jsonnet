local project = import 'brewkit/project.libsonnet';

local appIDs = [
    'productservice',
];

local proto = [
    'api/server/productinternal/productinternal.proto',
];

project.project(appIDs, proto)