local project = import 'brewkit/project.libsonnet';

local appIDs = [
    'userservice',
];

local proto = [
    'api/server/userinternal/userinternal.proto',
];

project.project(appIDs, proto)