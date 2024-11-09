local E = import 'extract.libsonnet';
local F = import 'fetch.libsonnet';
local P = import 'pack.libsonnet';
local S = import 'serve.libsonnet';
local V = import 'validate.libsonnet';
local W = import 'walk.libsonnet';

{
  // :: means "not visible in the output"
  EIGHT_SERVICES: {
    'user-provided': [
      E.cf,
      F.cf,
      P.cf,
      S.cf,
      V.cf,
      W.cf,
    ],
  },
}
