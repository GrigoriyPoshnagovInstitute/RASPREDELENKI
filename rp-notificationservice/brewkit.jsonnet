local project = import 'brewkit/project.libsonnet';

local appIDs = [
    'notificationservice',
];

local proto = [
    'api/server/notificationinternal/notificationinternal.proto',
];

project.project(appIDs, proto)