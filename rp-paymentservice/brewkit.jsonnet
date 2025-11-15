local project = import 'brewkit/project.libsonnet';

local appIDs = [
    'paymentservice',
];

local proto = [
    'api/server/paymentinternal/paymentinternal.proto',
];

project.project(appIDs, proto)