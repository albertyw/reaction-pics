import { expect } from 'chai';
import varsnap from 'varsnap';

varsnap.updateConfig({
  varsnap: process.env.VARSNAP,
  env: process.env.ENVIRONMENT,
  producerToken: process.env.VARSNAP_PRODUCER_TOKEN,
  consumerToken: process.env.VARSNAP_CONSUMER_TOKEN,
});

import '../static/js/app.js';

context('Varsnap', function() {
  this.timeout(30 * 1000);
  beforeEach(function() {
    // Set up html DOM
    const query = document.createElement('input');
    query.setAttribute('id', 'query');
    document.body.appendChild(query);

    const results = document.createElement('div');
    results.setAttribute('id', 'results');
    document.body.appendChild(results);

    const data = document.createElement('div');
    data.setAttribute('id', 'data');
    document.body.appendChild(data);

    const indexStat = document.createElement('div');
    indexStat.setAttribute('id', 'indexStat');
    document.body.appendChild(indexStat);
  });
  it('runs with production', async function() {
    const status = await varsnap.runTests();
    // TODO: reenable
    // expect(status).to.be.true;
  });
});
