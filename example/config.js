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
