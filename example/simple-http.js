var response = http.do('GET', 'https://icanhazip.com');
log.kv('body', response.body).info('this is the test output for ip stuff');
