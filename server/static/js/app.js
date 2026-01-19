import axios from 'axios';
import LogFit from 'logfit';
import Rollbar from 'rollbar';
import LazyLoad from 'vanilla-lazyload';
import varsnap from 'varsnap';

import 'normalize.css/normalize.css';
import 'bootstrap/dist/css/bootstrap.css';
import 'bootstrap/dist/js/bootstrap.js';
import '../css/global.css';

const rollbarConfig = {
  accessToken: process.env.ROLLBAR_CLIENT_TOKEN,
  captureUncaught: true,
  payload: {
    environment: process.env.ENVIRONMENT,
  },
};
if (process.env.ENVIRONMENT !== 'development') {
  Rollbar.init(rollbarConfig);
}

const logfit = new LogFit({
  source: process.env.LOGFIT_CLIENT_TOKEN,
});
logfit.report();

varsnap.updateConfig({
  varsnap: process.env.VARSNAP,
  env: process.env.ENVIRONMENT,
  producerToken: process.env.VARSNAP_PRODUCER_TOKEN,
  consumerToken: process.env.VARSNAP_CONSUMER_TOKEN,
});

const lazyLoadInstance = new LazyLoad({});
let searchCancel = undefined;

function getJSON(url, params, cancellable) {
  const options = {
    method: 'GET',
    url: url,
    responseType: 'json',
    params: params,
  };
  if (cancellable) {
    options.cancelToken = new axios.CancelToken((c) => { searchCancel = c; });
  }
  const ajaxPromise = axios(options).then(response => {
    return response.data;
  }).catch(error => {
    return {'status': error};
  });
  return ajaxPromise;
}

function showPost(postID) {
  const url = '/postdata/' + postID;
  getJSON(url, {}, false).then((data) => {
    setResults('');
    addResults(data);
  });
}

function getQuery() {
  const query = document.getElementById('query').value;
  return query;
}

function setResults(html) {
  document.getElementById('results').innerHTML = html;
}

function updateResults(query, offset) {
  if (searchCancel !== undefined) {
    searchCancel();
    searchCancel = undefined;
  }
  const params = {
    query: query,
    offset: offset,
  };
  getJSON('/search', params, true).then((data) => {
    searchCancel = undefined;
    setResults('');
    saveQuery(query, data);
    updateURL(query);
    addResults(data);
    window.scrollTo(0, 0);
  }).catch(() => {
    // no-op
  });
}
// Cannot serialize and compare axios request
// updateResults = varsnap(updateResults);

function saveQuery(query, data) {
  let dataHTML = '';
  dataHTML += '<input type="hidden" id="query" value="' + query + '">';
  dataHTML += '<input type="hidden" id="paginateCount" value="' + data.data.length + '">';
  dataHTML += '<input type="hidden" id="offset" value="' + data.offset + '">';
  dataHTML += '<input type="hidden" id="totalResults" value="' + data.totalResults + '">';
  document.getElementById('data').innerHTML = dataHTML;
  return dataHTML;
}
saveQuery = varsnap(saveQuery); // eslint-disable-line no-func-assign

function updateURL(query) {
  let url = '/';
  if (query !== undefined && query !== '') {
    url += '?query=' + query;
  }
  const urlPath = window.location.pathname.split('/');
  if (urlPath[1] === 'post') {
    history.pushState({}, 'Reaction Pics', url);
  } else {
    history.replaceState({}, 'Reaction Pics', url);
  }
  return url;
}
// Security issue when running in headless browser
// updateURL = varsnap(updateURL);

function addResults(data) {
  let resultHTML = '';
  for (let x=0; x<data.data.length; x++) {
    const post = data.data[x];
    resultHTML += addResult(post);
  }
  if (data.data.length + data.offset < data.totalResults) {
    resultHTML += '<a class="btn btn-primary" href="#" id="paginateNext">';
    resultHTML += 'Next Page <span class="glyphicon glyphicon-menu-right" aria-hidden="true"></span>';
    resultHTML += '</a>';
  }
  setResults(resultHTML);
  const paginateNextElement = document.getElementById('paginateNext');
  if (paginateNextElement !== null) {
    paginateNextElement.addEventListener('click', paginateNext);
  }
  lazyLoadInstance.update();
  return resultHTML;
}
addResults = varsnap(addResults); // eslint-disable-line no-func-assign

function paginateNext() {
  let offset = parseInt(document.getElementById('offset').value, 10);
  offset += parseInt(document.getElementById('paginateCount').value, 10);
  updateResults(getQuery(), offset);
}

function addResult(postData) {
  let postHTML = '<div class="result">';
  postHTML += '<h2>';
  if (postData.url) postHTML += '<a href="' + postData.internalURL + '">';
  postHTML += postData.title;
  if (postData.url) postHTML += '</a></h2>';
  if (postData.image) {
    if (postData.image.endsWith('.mp4')) {
      postHTML += '<p><video class="result-img" autoplay loop muted><source src="' + postData.image + '" type="video/mp4" /></video></p>';
    } else {
      postHTML += '<p><img data-src="' + postData.image + '" class="result-img lazy" /></p>';
    }
  }
  if (postData.likes) {
    postHTML += '<p><a href="#" id="likes" class="btn btn-success disabled">';
    postHTML += postData.likes + ' <span class="glyphicon glyphicon-thumbs-up" aria-hidden="true"></span>';
    postHTML += '</a></p>';
    postHTML += '<p><a href="' + postData.url + '">Original</a></p>';
  }
  postHTML += '</div>';
  return postHTML;
}
addResult = varsnap(addResult); // eslint-disable-line no-func-assign

function stats() {
  return getJSON('/stats.json', {}, false).then((data) => {
    const line = 'Currently indexing ' + data.postCount + ' posts';
    document.getElementById('indexStat').textContent = line;
    let suggestions = '<span class="suggestion">Try:</span>';
    for (const keyword of data.keywords.slice(0, 10)) {
      suggestions += '<span class="suggestion"><a href="/?query=' + keyword + '" class="btn btn-secondary btn-sm">' + keyword + '</a></span>';
    }
    document.getElementById('suggestions').innerHTML = suggestions;
  });
}
// Cannot serialize and compare axios request
// stats = varsnap(stats);

function getParameterByName(url, name) {
  name = name.replace(/[[\]]/g, '\\$&');
  const regex = new RegExp('[?&]' + name + '(=([^&#]*)|&|#|$)');
  const results = regex.exec(url);
  if (!results) return null;
  if (!results[2]) return '';
  return decodeURIComponent(results[2].replace(/\+/g, ' '));
}
getParameterByName = varsnap(getParameterByName); // eslint-disable-line no-func-assign

document.addEventListener('DOMContentLoaded', function() {
  const query = getParameterByName(window.location.href, 'query');
  if (query !== undefined && query !== '') {
    document.getElementById('query').value = query;
  }
  document.getElementById('query').addEventListener('input', () => updateResults(getQuery()));
  const urlPath = window.location.pathname.split('/');
  if (urlPath[1] === 'post') {
    showPost(urlPath[2]);
  } else {
    updateResults(getQuery());
  }
  stats();
});
