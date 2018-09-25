settings({
  name: 'Example Canary',
});

var frequency = '10m';
var timeout = '10m';

register_test(
  {
    name: 'simple example test',
    frequency: frequency,
    timeout: timeout,
  },
  file('simple-example.js')
);

register_test(
  {
    name: 'http demonstration',
    frequency: frequency,
    timeout: timeout,
  },
  file('simple-http.js')
);

register_test(
  {
    name: 'always pass',
    frequency: frequency,
    timeout: timeout,
  },
  file('always-pass.js')
);

register_test(
  {
    name: 'always fail',
    frequency: frequency,
    timeout: timeout,
  },
  file('always-fail.js')
);

register_test(
  {
    name: 'error thrown',
    frequency: frequency,
    timeout: timeout,
  },
  file('fail-error.js')
)
