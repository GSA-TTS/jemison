local B = import 'base.libsonnet';
local service = 'entree';

local credentials = [
  [
    'port',
    { cf: 8080, container: 8888, localhost: 8888 },
  ],
];

local parameters = [
  [
    'workers',
    { cf: 10, container: 50, localhost: 10},
  ],
  [
    'debug_level',
    { cf: 'warn', container: 'info', localhost: 'debug'},
  ],
] + B.parameters;

{
  creds:: [[service] + x for x in credentials],
  params:: [[service] + x for x in parameters],
  cf: { name: service } +
      B.params('credentials', 'cf', service, self.creds) +
      B.params('parameters', 'cf', service, self.params),
  container: { name: service } +
             B.params('credentials', 'container', service, self.creds) +
             B.params('parameters', 'container', service, self.params),
}
