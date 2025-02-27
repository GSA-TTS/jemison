local B = import 'base.libsonnet';
local service = 'extract';

local credentials = [
  [
    'port',
    { cf: 8080, container: 8888 },
  ],
];

local parameters = [
  [
    'workers',
    { cf: 10, container: 20 },
  ],
  [
    'debug_level',
    { cf: 'warn', container: 'warn' },
  ],
  [
    'extract_pdf',
    { cf: true, container: true },
  ],
  [
    'extract_html',
    { cf: true, container: true },
  ],
  [
    'walkabout',
    { cf: true, container: true },
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
