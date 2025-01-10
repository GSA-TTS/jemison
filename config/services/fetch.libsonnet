local B = import 'base.libsonnet';
local service = 'fetch';

local credentials = [
  [
    'port',
    { cf: 8080, container: 8888 },
  ],
];

local parameters = [
  // With the 'simple' queueing model, 1000+ workers is appropriate.
  // With 'round_robin', 10 is appropriate.
  // With 'one_per_domain', 10 is appropriate... but largely immaterial.
  [
    'queue_model',
    { cf: 'round_robin', container: 'round_robin' },
  ],
  [
    'workers',
    { cf: 10, container: 10 },
  ],
  [
    'debug_level',
    { cf: 'warn', container: 'info' },
  ],
  [
    'polite_sleep',
    { cf: 1, container: 1 },
  ],
  [
    'max_filesize_mb',
    { cf: 10, container: 20 },
  ],
  [
    'fetch_cooldown_ms',
    { cf: 500, container: 500 },
  ],
  [
    'fetch_poll_interval_ms',
    { cf: 1000, container: 1000 },
  ],
] + B.parameters;

{
  creds:: [[service] + x for x in credentials],
  params:: [[service] + x for x in parameters],
  cf: B.params('credentials', 'cf', service, self.creds) +
      B.params('parameters', 'cf', service, self.params),
  container: { name: service } +
             B.params('credentials', 'container', service, self.creds) +
             B.params('parameters', 'container', service, self.params),
}
