// const healthToggleBtn = document.getElementById('health-toggle');
const healthCheckConfigForm = document.getElementById('health-config-form');
const healthCheckConfigFailEvery = document.getElementById(
  'health-config-failevery'
);
const healthCheckConfigFailRatio = document.getElementById(
  'health-config-failratio'
);
const healthCheckConfigFailSeq = document.getElementById(
  'health-config-failseq'
);
const healthCheckConfigFailEveryBtn = document.getElementById(
  'health-config-failevery-set'
);
const healthCheckConfigFailRatioBtn = document.getElementById(
  'health-config-failratio-set'
);
const healthCheckConfigFailSeqBtn = document.getElementById(
  'health-config-failseq-set'
);

const healthStatusDiv = document.querySelector('.health-status');

const httpGetBtn = document.getElementById('http-get');
const httpUrlTxt = document.getElementById('http-url');
const httpBodyDiv = document.querySelector('.http-body');
const httpInfoDiv = document.querySelector('.http-info');
const httpRespStatusSpan = document.getElementById('http-resp-status')
const httpRespTookSpan = document.getElementById('http-resp-took')
const httpRespHeadersTable = document.getElementById('http-resp-headers')
const healthHistoryUl = document.getElementById('health-history-list');
const httpForm = document.getElementById('http-form');

var lastStatus = null;

function resetHealthcheckConfigForm() {
  fetch('/api/config/health?clear', {
    headers: new Headers({
      Authorization: 'Basic ' + btoa('username:password')
    })
  })
    .then(r => r.json())
    .then(d => {
      console.log(d);
      healthCheckConfigFailEvery.value = d['FailEvery'];
      healthCheckConfigFailRatio.value = d['FailRatio'];
      healthCheckConfigFailSeq.value = d['FailSeq'];
    });
}

function setHealthcheckConfig(key, value) {
  fetch(`/api/config/health?${key}=${value}`, {
    headers: new Headers({
      Authorization: 'Basic ' + btoa('username:password')
    })
  })
    .then(r => r.json())
    .then(d => {
      console.log(d);
      healthCheckConfigFailEvery.value = d['FailEvery'];
      healthCheckConfigFailRatio.value = d['FailRatio'];
      healthCheckConfigFailSeq.value = d['FailSeq'];
    });
}

function addHealthHistory(status) {
  let maxElements = 10;
  var li = document.createElement('li');
  if (status) {
    li.innerText = `✓ (${timestamp()})`;
    li.className = 'health-history-up';
  } else {
    li.innerText = `✗ (${timestamp()})`;
    li.className = 'health-history-down';
  }
  if (healthHistoryUl.children.length >= maxElements) {
    healthHistoryUl.removeChild(healthHistoryUl.children[maxElements - 1]);
  }

  healthHistoryUl.insertBefore(li, healthHistoryUl.children[0]);
}

function setHealthStatus(status) {
  let cl = healthStatusDiv.classList;
  let statusTxt = healthStatusDiv.children[0];

  if (status) {
    cl.remove('down');
    cl.add('up');
    statusTxt.innerText = `UP ✓ (last check: ${timestamp()})`;
  } else {
    cl.remove('up');
    cl.add('down');
    statusTxt.innerText = `DOWN ✗ (last check: ${timestamp()})`;
  }
  if (lastStatus != status) {
    addHealthHistory(status);
  }
}

function timestamp() {
  return new Date().toLocaleTimeString();
}

function updateHealthStatus() {
  let cl = healthStatusDiv.classList;
  let status = healthStatusDiv.children[0];
  fetch('/api/health').then(response => {
    setHealthStatus(response.ok);
    lastStatus = response.ok;
  });
}

function performGetRequest() {}

function updateHealthConfig() {
  fetch(
    '/api/config/health',
    { 
      headers: new Headers({
        'Authorization': 'Basic '+btoa('username:password'), 
      }), 
    }
  ).then(response => response.json())
  .then((conf) => {
    if (healthCheckConfigFailSeq !== document.activeElement) {
      healthCheckConfigFailSeq.value =  conf.FailSeq
    }
  }
  );

}

function updateUI() {
  updateHealthConfig();
  updateHealthStatus();
}

function toggleHealthStatus() {
  let cl = healthStatusDiv.classList;
  let p = healthStatusDiv.children[0];

  if (cl.contains('up')) {
    cl.remove('up');
    cl.add('down');
    p.innerText = 'DOWN ✗';
    return;
  }
  if (cl.contains('down')) {
    cl.remove('down');
    cl.add('up');
    p.innerText = 'UP ✓';
  }
}

// healthToggleBtn.addEventListener('click', function(e) {
//   toggleHealthStatus();
// });

function checkStatus(status) {
  if (status < 100) {
    return false
  }

  let hundreds = d => Math.floor(d/100)
  if (hundreds(status) == 5 || hundreds(status) == 4)  {
    return false
  }

  return true
}

function escapeHtml(unsafe) {
  return unsafe
       .replace(/&/g, "&amp;")
       .replace(/</g, "&lt;")
       .replace(/>/g, "&gt;")
       .replace(/"/g, "&quot;")
       .replace(/'/g, "&#039;");
}

console.log(httpForm);
httpForm.addEventListener('submit', function(e) {
  e.preventDefault();

  if (httpUrlTxt.value === '') {
    return;
  }
  httpUrlTxt.disabled = true;
  fetch(
    `/api/http/get?url=${encodeURI(httpUrlTxt.value)}`,
  ).then((response) => {
    return response.json()
  })
  .then((data) => {
    httpUrlTxt.disabled = false;
    console.log(data)
    httpInfoDiv.style.removeProperty('display');
    if (data.ServerStatus != "") {
      httpRespStatusSpan.innerText = `server error: ${data.ServerStatus}`
    } else {
      httpRespStatusSpan.innerText = `HTTP${data.ResponseStatus}`
    }
    
    httpRespTookSpan.innerText = `${data.TookMs} ms`
    
    var empty = document.createElement('tbody');
    var tbody = httpRespHeadersTable.firstChild
    httpRespHeadersTable.replaceChild(empty, tbody)

    for(const h in data.Headers) {
      var tr = httpRespHeadersTable.insertRow();
      var n = tr.insertCell();
      var v = tr.insertCell();
      n.innerText = h;
      v.innerText = data.Headers[h].join('/');

      console.log(tr);
    }
    httpBodyDiv.innerText = data.Body;
  }
  )
  .catch((error) => {
    console.log(`!!!ERROR: ${error}`)
  })

  
});

healthCheckConfigForm.addEventListener('submit', function(e) {
  console.log('submit');
  e.preventDefault();
});

healthCheckConfigForm.addEventListener('reset', function(e) {
  console.log('reset');
  e.preventDefault();
  resetHealthcheckConfigForm();
});

healthCheckConfigFailEveryBtn.addEventListener('click', function(e) {
  console.log(`FailEvery: ${healthCheckConfigFailEvery.value}`);
  setHealthcheckConfig('failevery', healthCheckConfigFailEvery.value);
});
healthCheckConfigFailRatioBtn.addEventListener('click', function(e) {
  console.log(`FailRatio: ${healthCheckConfigFailRatio.value}`);
  setHealthcheckConfig('failratio', healthCheckConfigFailRatio.value);
});
healthCheckConfigFailSeqBtn.addEventListener('click', function(e) {
  console.log(`FailSeq: ${healthCheckConfigFailSeq.value}`);
  setHealthcheckConfig('failseq', healthCheckConfigFailSeq.value);
});

updateUI();
window.setInterval(updateUI, 2000);
